package sqlLib

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB is a generic wrapper for database JSON/JSONB columns.
// It handles SQL NULLs via the Valid flag and remains transparent to REST APIs.
type JSONB[T any] struct {
	Data  T
	Valid bool // true if Data is present, false if NULL
}

// NewJSONB is a helper to quickly create a valid JSONB object.
// Example: models.NewJSONB([]string{"test"})
func NewJSONB[T any](data T) JSONB[T] {
	return JSONB[T]{
		Data:  data,
		Valid: true,
	}
}

// --- DATABASE INTERFACES ---

// Scan reads from the Database (implements sql.Scanner)
func (j *JSONB[T]) Scan(value interface{}) error {
	if value == nil {
		j.Data = *new(T) // Reset to zero-value
		j.Valid = false
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("JSONB scan: expected []byte")
	}
	j.Valid = true
	return json.Unmarshal(b, &j.Data)
}

// Value writes to the Database (implements driver.Valuer)
func (j JSONB[T]) Value() (driver.Value, error) {
	if !j.Valid {
		return nil, nil // Saves as SQL NULL
	}
	return json.Marshal(j.Data)
}

// --- REST API INTERFACES (The Magic) ---

// MarshalJSON ensures your API output looks normal (hides the wrapper)
func (j JSONB[T]) MarshalJSON() ([]byte, error) {
	if !j.Valid {
		return []byte("null"), nil // Outputs JSON null if invalid
	}
	return json.Marshal(j.Data)
}

// UnmarshalJSON ensures your API input binds normally (hides the wrapper)
func (j *JSONB[T]) UnmarshalJSON(data []byte) error {
	// If the frontend explicitly sends "null", mark as invalid
	if string(data) == "null" {
		j.Data = *new(T)
		j.Valid = false
		return nil
	}

	err := json.Unmarshal(data, &j.Data)
	if err == nil {
		j.Valid = true // Successfully parsed real data
	}
	return err
}
