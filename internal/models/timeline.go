package models

import (
	"encoding/json"
	"time"

	"github.com/bwmarrin/snowflake"
)

// represents a historical record of code execution
type Timeline struct {
	ID                 snowflake.ID  `json:"id" db:"id"`
	ProblemID          snowflake.ID  `json:"problemId" db:"problem_id"`
	InterviewSessionID *snowflake.ID `json:"interviewSessionId,omitempty" db:"interview_session_id"`
	SubmissionID       *snowflake.ID `json:"submissionId" db:"submission_id"`
	CodeSnapshotID     *snowflake.ID `json:"codeSnapshotId" db:"code_snapshot_id"`
	CreatedAt          time.Time     `json:"createdAt" db:"created_at"`
}

type DetailedTimeline struct {
	// Timeline info
	TimelineID         int64     `json:"timelineId" db:"timeline_id"`
	TimelineCreatedAt  time.Time `json:"timelineCreatedAt" db:"timeline_created_at"`
	InterviewSessionID *int64    `json:"interviewSessionId,omitempty" db:"interview_session_id"`

	// Code snapshot info (can be nil if only submission)
	CodeSnapshotID    *int64     `json:"codeSnapshotId,omitempty" db:"code_snapshot_id"`
	Code              *string    `json:"code,omitempty" db:"code"`
	SnapshotLanguage  *string    `json:"snapshotLanguage,omitempty" db:"snapshot_language"`
	SnapshotCreatedAt *time.Time `json:"snapshotCreatedAt,omitempty" db:"snapshot_created_at"`

	// Submission info (can be nil if only snapshot)
	SubmissionID        *int64          `json:"submissionId,omitempty" db:"submission_id"`
	SubmissionStatus    *string         `json:"submissionStatus,omitempty" db:"submission_status"`
	ExecutionTimeMs     *int            `json:"executionTimeMs,omitempty" db:"execution_time_ms"`
	MemoryUsedKb        *int            `json:"memoryUsedKb,omitempty" db:"memory_used_kb"`
	Stdin               *string         `json:"stdin,omitempty" db:"stdin"`
	Stdout              *string         `json:"stdout,omitempty" db:"stdout"`
	Stderr              *string         `json:"stderr,omitempty" db:"stderr"`
	TestCases           json.RawMessage `json:"testCases,omitempty" db:"test_cases"` // TODO: CHANGE THIS FROM RAW TO ACTUAL
	TestResults         json.RawMessage `json:"testResults,omitempty" db:"test_results"`
	SubmissionCreatedAt *time.Time      `json:"submissionCreatedAt,omitempty" db:"submission_created_at"`
}
