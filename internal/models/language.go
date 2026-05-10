package models

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

type Language struct {
	ID        snowflake.ID `json:"id" db:"id"`
	Name      string       `json:"name" db:"name"`
	Code      string       `json:"code" db:"code"`
	Extension string       `json:"extension" db:"extension"`
	CreatedAt time.Time    `json:"createdAt" db:"created_at"`
}
