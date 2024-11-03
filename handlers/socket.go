package handlers

import (
	"bruce/config"
	"bruce/handlers/queue"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coder/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type SocketMessage struct {
	MsgType  string `json:"MsgType"`
	Action   string `json:"Action"`
	ActionId string `json:"ActionId"`
	Message  string `json:"Message"`
}

type AuthMessage struct {
	MsgType string `json:"MsgType"`
	Id      string `json:"id"`
	Key     string `json:"key"`
}

func RetrieveEvents(item string, execution []config.Execution) (config.Execution, error) {
	for _, e := range execution {
		if e.Action == item {
			return e, nil
		}
	}
	return config.Execution{}, fmt.Errorf("execution %s not found", item)
}

// DataHandler processes messages sent over the WebSocket connection
func DataHandler(ctx context.Context, conn *websocket.Conn, eventExecutions []config.Execution) error {
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
				log.Debug().Err(err).Msg("DataHandler read error, likely connection lost")
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
			case "heartbeat":
				log.Debug().Msg("Sending heartbeat")
				d, err := json.Marshal(&SocketMessage{MsgType: "heartbeat", Message: "pong"})
				if err != nil {
					log.Error().Err(err).Msg("DataHandler failed to read heartbeat message")
					continue
				}
				queue.Add(d)
			case "execute":
				log.Debug().Msgf("Execute request received: %s", msg.Action)
				// Match the action with the execution in config
				log.Debug().Msgf("Action: %s, Execution: %#v", msg.Action, eventExecutions)
				actionEvent, err := RetrieveEvents(msg.Action, eventExecutions)
				if err != nil {
					log.Error().Err(err).Msgf("no such action: %s", msg.Action)
					sendMessage("execute-failure", fmt.Sprintf("no such action: %s", msg.Action), msg.Action, msg.ActionId)
					continue
				}
				// Execute the steps for the corresponding action
				t, err := config.LoadConfig(actionEvent.Target)
				if err != nil {
					log.Error().Err(err).Msgf("Cannot continue without configuration data, bad event config for: %s", actionEvent.Target)
					sendMessage("execute-failure", fmt.Sprintf("Cannot continue without configuration data, bad event action for: %s", actionEvent.Target), msg.Action, msg.ActionId)
					continue
				}
				err = ExecuteSteps(t)
				if err != nil {
					log.Error().Err(err).Msg("ExecuteSteps error")
					sendMessage("execute-failure", fmt.Sprintf("ExecuteSteps error: %s", err.Error()), msg.Action, msg.ActionId)
				}
				sendMessage("execute-success", "Execution completed", msg.Action, msg.ActionId)

			default:
				log.Warn().Msgf("DataHandler: Unknown message type: %s", msg.MsgType)
			}
		}
	}
}

func sendMessage(sub, body, action, actionId string) {
	authMsg := &SocketMessage{MsgType: sub, Message: body, Action: action, ActionId: actionId}
	d, err := json.Marshal(authMsg)
	if err != nil {
		log.Fatal().Err(err).Msg("DataHandler failed to encode basic message")
	}
	queue.Add(d) // Queue the message to send it
}

// SocketRunner: Handles the connection and initializes DataHandler
func SocketRunner(ctx context.Context, sockloc, skey, authkey string, eventExecutions []config.Execution) {
	// Initialize connection to WebSocket
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("SocketRunner received shutdown signal, exiting...")
			return // Exit the loop and stop reconnecting

		default:
			log.Info().Msgf("SocketRunner load balancing request on: %s", sockloc)
			socOpts := &websocket.DialOptions{
				HTTPHeader: http.Header{
					"Authorization": {fmt.Sprintf("%s:%s", skey, authkey)},
				},
			}
			c, _, err := websocket.Dial(ctx, sockloc, socOpts)
			if err != nil {
				log.Error().Err(err).Msg("SocketRunner failed to connect")
				time.Sleep(5 * time.Second) // Wait before retrying connection
				continue
			}

			log.Debug().Msg("SocketRunner connected successfully")

			// Start the DataHandler with the connection
			err = DataHandler(ctx, c, eventExecutions)
			if err != nil {
				log.Debug().Err(err).Msg("DataHandler error, connection likely lost")
				c.Close(websocket.StatusNormalClosure, "Connection lost, retrying...")
			}

			// Close the WebSocket connection and retry after 5 seconds
			c.Close(websocket.StatusNormalClosure, "Closing connection")
			log.Debug().Msg("SocketRunner connection closed...")
			time.Sleep(5 * time.Second)
		}
	}
}
