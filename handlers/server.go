package handlers

import (
	"bruce/config"
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
)

type SocketMessage struct {
	MsgType string          `json:"MsgType"`
	RawMsg  json.RawMessage `json:"-"`
}

type AuthMessage struct {
	MsgType string `json:"MsgType"`
	Id      string `json:"id"`
	Key     string `json:"key"`
}
type StatusMessage struct {
	MsgType string `json:"MsgType"`
	Message string `json:"sMessage"`
}

var inProgress bool

func RunServer(svr_config string, portNumber int) error {
	log.Debug().Msg("starting server task")

	// Read server configuration
	sc := &config.ServerConfig{}
	err := config.ReadServerConfig(svr_config, sc)
	if err != nil {
		log.Error().Err(err).Msg("cannot continue without configuration data")
		os.Exit(1)
	}

	// Channel to receive OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// WaitGroup to manage the lifecycle of goroutines
	var wg sync.WaitGroup

	// Context to manage cancellation of goroutines
	ctx, cancel := context.WithCancel(context.Background())

	log.Info().Msg("Starting Bruce in server mode")

	if len(sc.Execution) == 0 {
		log.Error().Msg("no execution targets configured")
		os.Exit(1)
	}

	for _, e := range sc.Execution {
		// Validate execution type
		if e.Type != "event" && e.Type != "cadence" {
			log.Info().Msgf("Skipping invalid execution target: %s must be of type 'event' or 'cadence'", e.Name)
			continue
		}

		// Start CadenceRunner
		if e.Type == "cadence" {
			wg.Add(1)
			go func(e config.Execution) {
				defer wg.Done()
				CadenceRunner(ctx, e.Name, e.Target, e.Cadence)
			}(e)
		}

		// Start SocketRunner
		if e.Type == "event" {
			wg.Add(1)
			go func(e config.Execution) {
				defer wg.Done()
				SocketRunner(ctx, e.Name, e.Target, sc.Endpoint, sc.Origin, sc.Key, e.Authorization)
			}(e)
		}
	}

	// Wait for a signal to shut down
	<-sigCh
	log.Info().Msg("Shutting down server...")

	// Cancel all goroutines
	cancel()

	// Wait for all runners to finish
	wg.Wait()

	log.Info().Msg("All runners finished, server shut down successfully.")
	return nil
}

func CadenceRunner(ctx context.Context, name, propfile string, cadence int) {
	log.Debug().Msgf("Starting CadenceRunner[%s] with propfile: %s, every %d minutes", name, propfile, cadence)
	t, err := config.LoadConfig(propfile)
	if err != nil {
		log.Error().Err(err).Msgf("cannot continue without configuration data, runner %s failed", name)
		return
	}

	// Run the task at the specified cadence interval
	ticker := time.NewTicker(time.Duration(cadence) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("CadenceRunner[%s] received cancellation signal, exiting...", name)
			return
		case <-ticker.C:
			log.Debug().Msgf("CadenceRunner[%s] running execution steps", name)
			err = ExecuteSteps(t)
			if err != nil {
				log.Error().Err(err).Msgf("CadenceRunner[%s] failed", name)
				return
			}
			log.Info().Msgf("CadenceRunner[%s] execution succeeded", name)
		}
	}
}

func SocketRunner(ctx context.Context, name, propfile, sockloc, origin, skey, authkey string) {
	log.Debug().Msgf("Starting SocketRunner[%s] with propfile: %s, socket: %s", name, propfile, sockloc)
	t, err := config.LoadConfig(propfile)
	if err != nil {
		log.Error().Err(err).Msgf("cannot continue without configuration data, runner %s failed", name)
		return
	}

	var c *websocket.Conn
	isConnected := false

	// Connection logic
	connect := func() error {
		var err error
		c, err = websocket.Dial(sockloc, "", origin)
		if err != nil {
			log.Error().Err(err).Msgf("SocketRunner[%s] cannot connect to socket: %s", name, sockloc)
			return err
		}
		isConnected = true
		log.Info().Msgf("SocketRunner[%s] successfully connected to socket: %s", name, sockloc)
		return nil
	}

	// Attempt initial connection
	if err := connect(); err != nil {
		// If initial connection fails, enter reconnect loop
		log.Info().Msgf("SocketRunner[%s] connection failed, attempting reconnect in 30s...", name)
		for {
			select {
			case <-ctx.Done():
				log.Info().Msgf("SocketRunner[%s] received cancellation signal, exiting...", name)
				return
			case <-time.After(30 * time.Second):
				log.Info().Msgf("SocketRunner[%s] retrying connection...", name)
				if err := connect(); err == nil {
					break
				}
			}
		}
	}

	// Close the connection cleanly when exiting
	defer func() {
		if isConnected {
			if err := c.Close(); err != nil {
				log.Error().Err(err).Msgf("SocketRunner[%s] error closing socket connection", name)
			} else {
				log.Info().Msgf("SocketRunner[%s] connection closed cleanly", name)
			}
		}
	}()

	// Main loop to listen for messages or handle reconnection
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("SocketRunner[%s] received cancellation signal, exiting...", name)
			return
		default:
			if !isConnected {
				log.Warn().Msgf("SocketRunner[%s] is disconnected, retrying connection...", name)
				for {
					select {
					case <-ctx.Done():
						log.Info().Msgf("SocketRunner[%s] received cancellation signal, exiting...", name)
						return
					case <-time.After(30 * time.Second):
						log.Info().Msgf("SocketRunner[%s] retrying connection...", name)
						if err := connect(); err == nil {
							break
						}
					}
				}
			}

			// Receive messages from the WebSocket connection
			var data []byte
			err := websocket.Message.Receive(c, &data)
			if err != nil {
				log.Error().Err(err).Msgf("SocketRunner[%s] lost connection to socket: %s, will attempt to reconnect...", name, sockloc)
				isConnected = false
				c.Close()
				continue // Enter reconnect loop
			}

			// Handle the received message
			msg := &SocketMessage{}
			err = json.Unmarshal(data, msg)
			if err != nil {
				log.Warn().Msgf("SocketRunner[%s] received non-application message: %s", name, string(data))
				continue
			}

			log.Info().Msgf("SocketRunner[%s] received valid message: %#v", name, msg)
			switch msg.MsgType {
			case "Execute":
				err = ExecuteSteps(t)
				if err != nil {
					log.Error().Err(err).Msgf("SocketRunner[%s] failed to execute steps", name)
					return
				}
				log.Info().Msgf("SocketRunner[%s] execution succeeded", name)
			case "Authenticate":
				log.Info().Msg("auth request received... sending authentication...")
				smsg := &AuthMessage{Id: skey, Key: authkey, MsgType: "Authenticate"}
				d, err := json.Marshal(smsg)
				if err != nil {
					log.Error().Err(err).Msg("could not marshall auth keys")
					continue
				}
				c.Write(d)
			case "AuthResult":
				msg := &StatusMessage{}
				err := json.Unmarshal(data, msg)
				if err != nil {
					log.Error().Err(err).Msg("could not read status message")
					continue
				}
				log.Info().Msgf("Auth status result: %s", msg.Message)
			default:
				log.Warn().Msgf("SocketRunner[%s] received unknown message type: %s", name, msg.MsgType)
			}
		}
	}
}

func ExecuteSteps(t *config.TemplateData) error {
	for idx, step := range t.Steps {
		if step.Action != nil {
			err := step.Action.Execute()
			if err != nil {
				log.Error().Err(err).Msgf("error executing step [%d]", idx+1)
				return err
			}
		}
	}
	return nil
}
