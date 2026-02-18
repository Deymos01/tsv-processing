package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/Deymos01/tsv-processing/internal/domain"
	"github.com/Deymos01/tsv-processing/internal/transport/http/dto"
	"github.com/Deymos01/tsv-processing/internal/usecase"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 1000
)

// MessageHandler handles HTTP requests related to messages.
type MessageHandler struct {
	messageUC usecase.MessageUseCase
	log       *zap.Logger
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(messageUC usecase.MessageUseCase, log *zap.Logger) *MessageHandler {
	return &MessageHandler{
		messageUC: messageUC,
		log:       log,
	}
}

// GetByUnitGUID handles GET /api/v1/messages?...
//
// Query params:
//   - unit_guid (required) — filter by device GUID
//   - page      (optional, default 1)
//   - limit     (optional, default 20, max 1000)
func (h *MessageHandler) GetByUnitGUID(w http.ResponseWriter, r *http.Request) {
	unitGUID := r.URL.Query().Get("unit_guid")
	if unitGUID == "" {
		h.writeError(w, http.StatusBadRequest, "unit_guid query parameter is required")
		return
	}

	page := parseIntParam(r.URL.Query().Get("page"), defaultPage)
	limit := parseIntParam(r.URL.Query().Get("limit"), defaultLimit)

	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	filter := domain.MessageFilter{
		UnitGUID: unitGUID,
		Page:     page,
		Limit:    limit,
	}

	messages, total, err := h.messageUC.GetByUnitGUID(r.Context(), filter)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			h.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.log.Error("failed to get messages",
			zap.String("unit_guid", unitGUID),
			zap.Error(err),
		)
		h.writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := dto.ToMessageListResponse(messages, total, page, limit)
	h.writeJSON(w, http.StatusOK, resp)
}

// writeJSON encodes v as JSON and writes it to the response.
func (h *MessageHandler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.log.Error("failed to encode JSON response", zap.Error(err))
	}
}

// writeError writes a standard JSON error response.
func (h *MessageHandler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, dto.NewErrorResponse(msg))
}

// parseIntParam parses a string query param as int, returning fallback on failure.
func parseIntParam(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
