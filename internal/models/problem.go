package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
	sqlLib "github.com/himanshu3889/code-master-backend/base/lib/sql"
)

// SavedCode maps LanguageCode -> Code string
// Example: {"python": "print(1)", "go": "fmt.Println(1)"}
type SavedCode map[string]string

// ProblemTestCase represents a single test case for the problem
type ProblemTestCase struct {
	Input  string `json:"input" db:"input"`
	Output string `json:"output" db:"output"`
}

// Problem represents a code problem
type Problem struct {
	ID              snowflake.ID                    `json:"id" db:"id"`
	Name            string                          `json:"name" db:"name"`
	Group           string                          `json:"group" db:"group_name"`
	Description     string                          `json:"description" db:"description"`
	DifficultyLevel *string                         `json:"difficultyLevel" db:"difficulty_level"`
	Rating          *int                            `json:"rating" db:"rating"`
	TimeLimit       *int                            `json:"timeLimit" db:"time_limit_ms"`
	MemoryLimit     *int                            `json:"memoryLimit" db:"memory_limit_mb"`
	TestCases       sqlLib.JSONB[[]ProblemTestCase] `json:"testCases" db:"test_cases"`
	RawPayload      string                          `json:"-" db:"raw_payload"`
	URL             string                          `json:"url,omitempty" db:"url"`
	Status          string                          `json:"status" db:"status"`
	SavedCode       sqlLib.JSONB[SavedCode]         `json:"savedCode" db:"saved_code"`
	AttemptCount    int                             `json:"attemptCount" db:"attempt_count"`
	Notes           string                          `json:"notes" db:"notes"`
	CreatedAt       time.Time                       `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time                       `json:"updatedAt" db:"updated_at"`
}
