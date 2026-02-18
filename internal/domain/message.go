package domain

import (
	"time"
)

// Message represents a single parsed row from a TSV input file.
type Message struct {
	ID                  int64
	Number              int
	MQTT                string
	InvID               string
	UnitGUID            string
	MessageID           string
	MessageText         string
	MessageContext      string
	MessageClass        string
	MessageLevel        string
	VariableZone        string
	VariableAddress     string
	UseAsBlockStart     bool
	Type                string
	BitNumberInRegister int
	InvertBit           bool
	SourceFile          string
	CreatedAt           time.Time
}

// MessageFilter defines query parameters for fetching messages.
type MessageFilter struct {
	UnitGUID string
	Page     int
	Limit    int
}

// Offset calculates the SQL OFFSET value from page and limit.
func (f MessageFilter) Offset() int {
	if f.Page <= 1 {
		return 0
	}
	return (f.Page - 1) * f.Limit
}
