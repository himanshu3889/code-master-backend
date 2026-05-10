package appWebsocket

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/himanshu3889/code-master-backend/internal/models"
	"github.com/sirupsen/logrus"
)

// Handle submission message
func HandleCodeSubmissionResults(submission *models.Submission) {
	// Ensure the client actually exists
	if userClient == nil {
		logrus.Info("Cannot send result: userClient is nil (user likely disconnected)")
		return
	}
	if userClient.send == nil {
		logrus.Info("Cannot send result: userClient.send channel is nil")
		return
	}

	// Make submission websocket message
	submissionBytes, err := json.Marshal(submission)
	if err != nil {
		return
	}

	submissionRawMsg := json.RawMessage(submissionBytes)

	submissionMsg := &SocketMessage{
		Event: EventSubmissionResult,
		Room:  fmt.Sprintf("problem.%s", submission.ProblemID),
		Data:  &submissionRawMsg,
	}

	// Prepare the message
	messageBytes, err := json.Marshal(submissionMsg)
	if err != nil {
		return
	}
	preparedMsg, err := websocket.NewPreparedMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		return
	}

	// Send to client via channel
	userClient.send <- preparedMsg
}

// Handle new problem in system
func HandleNewProblem(problem *models.Problem) {
	// Ensure the client actually exists
	if userClient == nil {
		logrus.Info("Cannot send result: userClient is nil (user likely disconnected)")
		return
	}
	if userClient.send == nil {
		logrus.Info("Cannot send result: userClient.send channel is nil")
		return
	}

	// Make problem websocket message
	problemBytes, err := json.Marshal(problem)
	if err != nil {
		return
	}

	problemRawMsg := json.RawMessage(problemBytes)

	submissionMsg := &SocketMessage{
		Event: EventNewProblem,
		Room:  "any",
		Data:  &problemRawMsg,
	}

	// Prepare the message
	messageBytes, err := json.Marshal(submissionMsg)
	if err != nil {
		return
	}
	preparedMsg, err := websocket.NewPreparedMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		return
	}

	// Send to client via channel
	userClient.send <- preparedMsg
}
