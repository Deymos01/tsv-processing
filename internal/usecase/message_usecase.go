package usecase

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Deymos01/tsv-processing/internal/domain"
	"github.com/Deymos01/tsv-processing/internal/repository"
)

// messageUseCase implements MessageUseCase.
type messageUseCase struct {
	messageRepo repository.MessageRepository
	log         *zap.Logger
}

// NewMessageUseCase constructs a new messageUseCase.
func NewMessageUseCase(repo repository.MessageRepository, log *zap.Logger) MessageUseCase {
	return &messageUseCase{
		messageRepo: repo,
		log:         log,
	}
}

// GetByUnitGUID validates the filter and delegates to the repository.
func (uc *messageUseCase) GetByUnitGUID(
	ctx context.Context,
	filter domain.MessageFilter,
) ([]domain.Message, int, error) {
	if filter.UnitGUID == "" {
		return nil, 0, fmt.Errorf("get messages: %w", domain.ErrInvalidInput)
	}
	if filter.Limit <= 0 {
		filter.Limit = 20 // sensible default
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	messages, total, err := uc.messageRepo.GetByUnitGUID(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("get messages by unit_guid %q: %w", filter.UnitGUID, err)
	}

	uc.log.Debug("fetched messages by unit_guid",
		zap.String("unit_guid", filter.UnitGUID),
		zap.Int("count", len(messages)),
		zap.Int("total", total),
	)

	return messages, total, nil
}
