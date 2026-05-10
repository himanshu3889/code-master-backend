package appWebsocket

import (
	"context"
	"encoding/json"

	"github.com/bwmarrin/snowflake"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/sirupsen/logrus"
)

type EventType string

// Last after . is the action

const (
	// Snapshot Event
	EventCodeSnapshotMessageSubmit EventType = "codeSnapshot.submit"
	// Submission result
	EventSubmissionResult EventType = "submission.result"
	// Problem code save
	EventProblemCodeSave EventType = "problem.code.save"
	// New problem in system
	EventNewProblem EventType = "problem.new"
)

type SocketMessage struct {
	Event EventType        `json:"event"`
	Room  string           `json:"room"`
	Data  *json.RawMessage `json:"data"`
}

// Handle the incoming message from the user
func (client *Client) handleIncomingMessage(recMessage []byte) {
	var msg SocketMessage
	if err := json.Unmarshal(recMessage, &msg); err != nil {
		// logrus.WithError(err).Warn("Invalid message format")
		return
	}

	// logrus.Infof("Received: %s", recMessage)
	if msg.Room == "" {
		logrus.Warn("Missing room in message")
		return
	}

	switch msg.Event {
	case EventCodeSnapshotMessageSubmit:
		client.handleCodeSnapshotMessageSubmit(&msg)
	case EventProblemCodeSave:
		client.handleProblemCodeSave(&msg)
	default:
		logrus.Warnf("Unknown event '%s'", msg.Event)
	}
}

// Handle submission message
func (client *Client) handleCodeSnapshotMessageSubmit(msg *SocketMessage) {
	snapshot := models.CodeSnapshot{}
	if err := json.Unmarshal(*msg.Data, &snapshot); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal code snapshot")
		return
	}
	appErr := websocketStore.CreateCodeSnapshot(context.Background(), &snapshot)
	if appErr != nil {
		return
	}

}

// Define a temporary struct for the incoming message
type SaveCodePayload struct {
	ID           snowflake.ID `json:"id"`
	LanguageCode string       `json:"langCode"`
	Code         string       `json:"code"`
}

func (client *Client) handleProblemCodeSave(msg *SocketMessage) {
	var payload SaveCodePayload

	// Bind the specific fields from the JSON message
	if err := json.Unmarshal(*msg.Data, &payload); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal save code payload")
		return
	}

	// Pass the individual fields to the Store method we refactored earlier
	appErr := websocketStore.SaveProblemCode(
		context.Background(),
		payload.ID,
		payload.LanguageCode,
		payload.Code,
	)

	if appErr != nil {
		logrus.WithFields(logrus.Fields{
			"problem_id": payload.ID,
			"error":      appErr.Message,
		}).Error("Failed to save problem code via websocket")
		return
	} else {
		logrus.Infof("Problem %d code saved successfully", payload.ID)
	}
}
