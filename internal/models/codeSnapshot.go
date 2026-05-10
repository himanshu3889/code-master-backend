package models

import (
	"github.com/bwmarrin/snowflake"
	"time"
)

// represents a historical record of code execution
type CodeSnapshot struct {
	ID        snowflake.ID `json:"id" db:"id"`
	ProblemID snowflake.ID `json:"problemId" db:"problem_id"`
	Language  string       `json:"language" db:"language"`
	Code      string       `json:"code" db:"code"`
	CreatedAt time.Time    `json:"createdAt" db:"created_at"`
}
