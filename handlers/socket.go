package handlers

import (
	"bruce/config"
	"context"
	"encoding/json"
	"github.com/coder/websocket/wsjson"
	"time"

	"github.com/coder/websocket"
	"github.com/rs/zerolog/log"
)

const heartbeatInterval = 5 * time.Second
const reconnectInterval = 5 * time.Second
const connectionTimeout = 60 * time.Second

type SocketMessage struct {
	MsgType string `json:"MsgType"`
	Message string `json:"Message"`
}
type AuthMessage struct {
	MsgType string `json:"MsgType"`
	Id      string `json:"id"`
	Key     string `json:"key"`
}

func SocketRunner(ctx context.Context, name, propfile, sockloc, skey, authkey string) {
	log.Debug().Msgf("Starting SocketRunner[%s] with propfile: %s, socket: %s", name, propfile, sockloc)

	// Load configuration
	t, err := config.LoadConfig(propfile)
	if err != nil {
		log.Error().Err(err).Msgf("Cannot continue without configuration data, runner %s failed", name)
		return
	}
	isConnected := false
	var conn *websocket.Conn

	// Main loop for message handling and reconnection
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("SocketRunner[%s] received system exit signal, exiting...", name)
			if conn != nil {
				conn.CloseNow() // Close connection when shutting down
			}
			return
		case <-time.After(heartbeatInterval):
			// Send heartbeat if connected
			log.Info().Msgf("SocketRunner[%s] sending heartbeat...", name)
			err = wsjson.Write(ctx, conn, &SocketMessage{MsgType: "heartbeat", Message: "ping"})
			if err != nil {
				log.Error().Err(err).Msgf("SocketRunner[%s] failed to send heartbeat, marking as disconnected", name)
				isConnected = false
			}

		default:
			if !isConnected {
				log.Info().Msgf("SocketRunner[%s] attempting to connect...", name)
				conn, err = connectToSocket(ctx, sockloc)
				if err != nil {
					isConnected = false
					log.Error().Err(err).Msgf("SocketRunner[%s] failed to connect to socket: %s", name, sockloc)
					continue
				}
				isConnected = true
			}

			log.Info().Msgf("SocketRunner[%s] waiting for messages...", name)
			// Listen for messages while the connection is active
			mt, data, cerr := conn.Read(ctx)
			if cerr != nil {
				log.Error().Err(cerr).Msgf("SocketRunner[%s] read error", name)
				isConnected = false
				time.Sleep(reconnectInterval)
				continue
			}
			// Handle received message
			if mt == websocket.MessageText {
				err = handleMessage(data, t, name, skey, authkey, ctx, conn)
				if err != nil {
					log.Error().Err(err).Msgf("SocketRunner[%s] error handling message", name)
				}
			}

		}
	}
}

func connectToSocket(ctx context.Context, sockloc string) (*websocket.Conn, error) {
	conn, _, err := websocket.Dial(ctx, sockloc, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func handleMessage(data []byte, t *config.TemplateData, name, ident, key string, ctx context.Context, conn *websocket.Conn) error {
	// Parse the received message
	msg := &SocketMessage{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		log.Warn().Msgf("SocketRunner[%s] received non-application message: %s", name, string(data))
		return err
	}

	log.Info().Msgf("SocketRunner[%s] received valid message: %#v", name, msg)

	switch msg.MsgType {
	case "execute":
		// Handle execution command
		err = ExecuteSteps(t)
		if err != nil {
			log.Error().Err(err).Msgf("SocketRunner[%s] failed to execute steps", name)
			return err
		}
		log.Info().Msgf("SocketRunner[%s] execution succeeded", name)

	case "authenticate":
		// Handle authentication
		log.Info().Msgf("SocketRunner[%s] received authentication request", name)
		authMsg := &AuthMessage{Id: ident, Key: key, MsgType: "Authenticate"}
		d, err := json.Marshal(authMsg)
		if err != nil {
			log.Error().Err(err).Msgf("SocketRunner[%s] failed to marshal auth message", name)
			return err
		}
		err = wsjson.Write(ctx, conn, d)
		if err != nil {
			log.Error().Err(err).Msgf("SocketRunner[%s] failed to send auth message", name)
			return err
		}

	case "authresult":
		// Handle authentication result
		result := &SocketMessage{}
		err = json.Unmarshal(data, result)
		if err != nil {
			log.Error().Err(err).Msgf("SocketRunner[%s] failed to parse auth result", name)
			return err
		}
		log.Info().Msgf("SocketRunner[%s] authentication result: %s", name, result.Message)

	default:
		// Handle unknown message types
		log.Warn().Msgf("SocketRunner[%s] received unknown message type: %s", name, msg.MsgType)
	}
	return nil
}
