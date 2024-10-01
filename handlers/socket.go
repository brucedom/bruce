package handlers

import (
	"bruce/config"
	"bruce/handlers/queue"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coder/websocket"
	"github.com/rs/zerolog/log"
	"time"
)

type SocketMessage struct {
	MsgType string `json:"MsgType"`
	Message string `json:"Message"`
}

type AuthMessage struct {
	MsgType string `json:"MsgType"`
	Id      string `json:"id"`
	Key     string `json:"key"`
}

// DataHandler: Processes messages sent over the WebSocket connection
func DataHandler(ctx context.Context, conn *websocket.Conn, skey, authkey string, execution []config.Execution) error {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("DataHandler shutting down...")
			return nil
		default:
			// Process messages from the queue (if any) and send them
			if conn != nil && queue.HasMessages() {
				for queue.HasMessages() {
					msg := queue.GetNext() // Get the next message in the queue
					if msg != nil {
						err := conn.Write(ctx, websocket.MessageText, msg)
						if err != nil {
							log.Error().Err(err).Msg("DataHandler failed to write message from queue")
							queue.Add(msg) // Re-add the message to the queue if the write fails
							return err     // Return error to indicate connection loss
						} else {
							queue.Remove(msg) // Remove the message from the queue after successful write
						}
					}
				}
			}

			// Handle other messages coming from WebSocket
			_, data, err := conn.Read(ctx)
			if err != nil {
				log.Error().Err(err).Msg("DataHandler read error, likely connection lost")
				return err // Return error to indicate connection loss
			}

			// Process the received data
			msg := &SocketMessage{}
			err = json.Unmarshal(data, msg)
			if err != nil {
				log.Warn().Msgf("DataHandler received invalid message: %s", string(data))
				continue
			}

			// Handle message based on type
			switch msg.MsgType {
			case "authenticate":
				log.Info().Msg("Authentication requested ")
				// Prepare authentication message and send it directly to the WebSocket
				authMsg := &SocketMessage{MsgType: "authenticate", Message: fmt.Sprintf("%s:%s", skey, authkey)}
				d, err := json.Marshal(authMsg)
				if err != nil {
					log.Error().Err(err).Msg("DataHandler failed to read authentication message")
					continue
				}
				queue.Add(d) // Queue the message to send it
				log.Info().Msg("Authentication response queued...")
			case "heartbeat":
				log.Info().Msg("Sending heartbeat")
				d, err := json.Marshal(&SocketMessage{MsgType: "heartbeat", Message: "pong"})
				if err != nil {
					log.Error().Err(err).Msg("DataHandler failed to read heartbeat message")
					continue
				}
				queue.Add(d)
			case "authentication":
				log.Info().Msgf("Authentication response received: %s", msg.Message)
			case "execute":
				log.Info().Msgf("Execute request received: %s", msg.Message)
				// Match the action with the execution in config
				for _, exec := range execution {
					if exec.Action == msg.Message {
						// Execute the steps for the corresponding action
						t, err := config.LoadConfig(exec.Target)
						if err != nil {
							log.Error().Err(err).Msgf("Cannot continue without configuration data, bad config file or missing config file at: %s", exec.Target)
							continue
						}
						err = ExecuteSteps(t)
						if err != nil {
							log.Error().Err(err).Msg("ExecuteSteps error")
						}
					}
				}
			default:
				log.Warn().Msgf("DataHandler: Unknown message type: %s", msg.MsgType)
			}
		}
	}
}

// SocketRunner: Handles the connection and initializes DataHandler
func SocketRunner(ctx context.Context, sockloc, skey, authkey string, execution []config.Execution) {
	// Initialize connection to WebSocket
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("SocketRunner received shutdown signal, exiting...")
			return // Exit the loop and stop reconnecting

		default:
			log.Info().Msgf("SocketRunner attempting to connect to: %s", sockloc)
			c, _, err := websocket.Dial(ctx, sockloc, nil)
			if err != nil {
				log.Error().Err(err).Msg("SocketRunner failed to connect")
				time.Sleep(5 * time.Second) // Wait before retrying connection
				continue
			}

			log.Info().Msg("SocketRunner connected successfully")

			// Start the DataHandler with the connection
			err = DataHandler(ctx, c, skey, authkey, execution)
			if err != nil {
				log.Error().Err(err).Msg("DataHandler error, connection likely lost")
				c.Close(websocket.StatusNormalClosure, "Connection lost, retrying...")
			}

			// Close the WebSocket connection and retry after 5 seconds
			c.Close(websocket.StatusNormalClosure, "Closing connection")
			log.Info().Msg("SocketRunner closing connection, retrying...")
			time.Sleep(5 * time.Second)
		}
	}
}
