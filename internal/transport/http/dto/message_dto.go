package dto

import "github.com/Deymos01/tsv-processing/internal/domain"

// MessageResponse is the JSON representation of a single message.
type MessageResponse struct {
	ID                  int64  `json:"id"`
	Number              int    `json:"number"`
	MQTT                string `json:"mqtt"`
	InvID               string `json:"inv_id"`
	UnitGUID            string `json:"unit_guid"`
	MessageID           string `json:"message_id"`
	MessageText         string `json:"message_text"`
	MessageContext      string `json:"message_context"`
	MessageClass        string `json:"message_class"`
	MessageLevel        string `json:"message_level"`
	VariableZone        string `json:"variable_zone"`
	VariableAddress     string `json:"variable_address"`
	UseAsBlockStart     bool   `json:"use_as_block_start"`
	Type                string `json:"type"`
	BitNumberInRegister int    `json:"bit_number_in_register"`
	InvertBit           bool   `json:"invert_bit"`
	SourceFile          string `json:"source_file"`
	CreatedAt           string `json:"created_at"`
}

// MessageListResponse wraps a paginated list of messages.
type MessageListResponse struct {
	Data       []MessageResponse `json:"data"`
	Pagination PaginationMeta    `json:"pagination"`
}

// ToMessageResponse maps a domain.Message to a MessageResponse DTO.
func ToMessageResponse(m domain.Message) MessageResponse {
	return MessageResponse{
		ID:                  m.ID,
		Number:              m.Number,
		MQTT:                m.MQTT,
		InvID:               m.InvID,
		UnitGUID:            m.UnitGUID,
		MessageID:           m.MessageID,
		MessageText:         m.MessageText,
		MessageContext:      m.MessageContext,
		MessageClass:        m.MessageClass,
		MessageLevel:        m.MessageLevel,
		VariableZone:        m.VariableZone,
		VariableAddress:     m.VariableAddress,
		UseAsBlockStart:     m.UseAsBlockStart,
		Type:                m.Type,
		BitNumberInRegister: m.BitNumberInRegister,
		InvertBit:           m.InvertBit,
		SourceFile:          m.SourceFile,
		CreatedAt:           m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToMessageListResponse maps a slice of domain.Message to a paginated DTO.
func ToMessageListResponse(messages []domain.Message, total, page, limit int) MessageListResponse {
	data := make([]MessageResponse, 0, len(messages))
	for _, m := range messages {
		data = append(data, ToMessageResponse(m))
	}

	return MessageListResponse{
		Data:       data,
		Pagination: NewPaginationMeta(total, page, limit),
	}
}
