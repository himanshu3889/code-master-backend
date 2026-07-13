package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

type InterviewSessionStatus string

const (
	InterviewSessionInProgress InterviewSessionStatus = "IN_PROGRESS"
	InterviewSessionCompleted  InterviewSessionStatus = "COMPLETED"
	InterviewSessionTimeout    InterviewSessionStatus = "TIMEOUT"
	InterviewSessionAbandoned  InterviewSessionStatus = "ABANDONED"
)

type InterviewSession struct {
	ID               snowflake.ID           `json:"id" db:"id"`
	ProblemID        snowflake.ID           `json:"problemId" db:"problem_id"`
	Status           InterviewSessionStatus `json:"status" db:"status"`
	TimeLimitSeconds int                    `json:"timeLimitSeconds" db:"time_limit_seconds"`
	StartedAt        time.Time              `json:"startedAt" db:"started_at"`
	EndedAt          *time.Time             `json:"endedAt,omitempty" db:"ended_at"`
	AutoSubmitted    bool                   `json:"autoSubmitted" db:"auto_submitted"`
	CreatedAt        time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time              `json:"updatedAt" db:"updated_at"`
}
