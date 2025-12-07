package v1

import (
	"github.com/google/uuid"
)

// Result represents the result of a cart validation
type Result struct {
	Valid    bool
	Errors   []Error
	Warnings []Warning
}

// Error represents a validation error
type Error struct {
	GoodID  uuid.UUID
	Message string
	Code    string
}

// Warning represents a validation warning
type Warning struct {
	GoodID  uuid.UUID
	Message string
	Code    string
}

