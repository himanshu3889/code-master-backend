package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
	sqlLib "github.com/himanshu3889/code-master-backend/base/lib/sql"
)

// ExecutionStatus represents the status of a code execution
type ExecutionStatus string

const (
	StatusAccepted            ExecutionStatus = "AC"
	StatusCompilationError    ExecutionStatus = "CE"
	StatusTimeLimitExceeded   ExecutionStatus = "TLE"
	StatusMemoryLimitExceeded ExecutionStatus = "MLE"
	StatusWrongAnswer         ExecutionStatus = "WA"
	StatusRuntimeError        ExecutionStatus = "RE"
)

// TestResult represents the result of running a single test case
type TestResult struct {
	TestIndex   int             `json:"testIndex"`
	Passed      bool            `json:"passed"`
	Status      ExecutionStatus `json:"status"`
	TimeMs      int64           `json:"timeMs"`
	MemoryBytes int64           `json:"memoryBytes"`
	Stderr      *string         `json:"stderr,omitempty"`
	Stdout      *string         `json:"stdout,omitempty" db:"stdout"`
}

// represents a historical record of code execution
type Submission struct {
	ID                 snowflake.ID                `json:"id" db:"id"`
	ProblemID          snowflake.ID                `json:"problemId" db:"problem_id"`
	InterviewSessionId *snowflake.ID               `json:"interviewSessionId" db:"interview_session_id"`
	Language           string                      `json:"language" db:"language"`
	Status             *ExecutionStatus            `json:"status" db:"status"`
	ExecutionTimeMs    *int64                      `json:"executionTimeMs,omitempty" db:"execution_time_ms"`
	MemoryUsedKb       *int64                      `json:"memoryUsedKb,omitempty" db:"memory_used_kb"`
	Stdin              string                      `json:"stdin" db:"stdin"`
	Stdout             *string                     `json:"stdout,omitempty" db:"stdout"`
	Stderr             *string                     `json:"stderr,omitempty" db:"stderr"`
	TestCases          sqlLib.JSONB[[]string]      `json:"testCases" db:"test_cases"`
	TestResults        sqlLib.JSONB[[]*TestResult] `json:"testResults" db:"test_results"`
	CreatedAt          time.Time                   `json:"createdAt" db:"created_at"`
}
