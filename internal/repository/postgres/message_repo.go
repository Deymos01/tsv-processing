package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Deymos01/tsv-processing/internal/domain"
)

// MessageRepo implements repository.MessageRepository backed by PostgreSQL.
type MessageRepo struct {
	pool *pgxpool.Pool
}

// NewMessageRepo creates a new MessageRepo.
func NewMessageRepo(pool *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{pool: pool}
}

// BulkInsert saves a batch of messages using pgx CopyFrom.
// All rows are inserted in a single transaction; the entire batch is rolled back on error.
func (r *MessageRepo) BulkInsert(ctx context.Context, messages []domain.Message) error {
	if len(messages) == 0 {
		return nil
	}

	columns := []string{
		"number", "mqtt", "inv_id", "unit_guid", "message_id",
		"message_text", "context", "message_class", "message_level",
		"variable_zone", "variable_address", "use_as_block_start",
		"type", "bit_number_in_register", "invert_bit", "source_file",
	}

	rows := make([][]any, 0, len(messages))
	for _, m := range messages {
		rows = append(rows, []any{
			m.Number, m.MQTT, m.InvID, m.UnitGUID, m.MessageID,
			m.MessageText, m.MessageContext, m.MessageClass, m.MessageLevel,
			m.VariableZone, m.VariableAddress, m.UseAsBlockStart,
			m.Type, m.BitNumberInRegister, m.InvertBit, m.SourceFile,
		})
	}

	_, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"messages"},
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("bulk insert messages: %w", err)
	}

	return nil
}

// GetByUnitGUID returns paginated messages and the total count for a given unit_guid.
func (r *MessageRepo) GetByUnitGUID(
	ctx context.Context,
	filter domain.MessageFilter,
) ([]domain.Message, int, error) {
	const countQ = `SELECT COUNT(*) FROM messages WHERE unit_guid = $1`

	var total int
	if err := r.pool.QueryRow(ctx, countQ, filter.UnitGUID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count messages by unit_guid: %w", err)
	}

	if total == 0 {
		return []domain.Message{}, 0, nil
	}

	const q = `
		SELECT
			id, number, mqtt, inv_id, unit_guid, message_id,
			message_text, context, message_class, message_level,
			variable_zone, variable_address, use_as_block_start,
			type, bit_number_in_register, invert_bit, source_file, created_at
		FROM messages
		WHERE unit_guid = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, q, filter.UnitGUID, filter.Limit, filter.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("query messages by unit_guid: %w", err)
	}
	defer rows.Close()

	messages, err := scanMessages(rows)
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetDistinctUnitGUIDs returns all unique unit_guid values found in a given source file.
// Used after a successful insert to know which RTF files to generate.
func (r *MessageRepo) GetDistinctUnitGUIDs(ctx context.Context, sourceFile string) ([]string, error) {
	const q = `SELECT DISTINCT unit_guid FROM messages WHERE source_file = $1`

	rows, err := r.pool.Query(ctx, q, sourceFile)
	if err != nil {
		return nil, fmt.Errorf("get distinct unit_guids: %w", err)
	}
	defer rows.Close()

	var guids []string
	for rows.Next() {
		var g string
		if err = rows.Scan(&g); err != nil {
			return nil, fmt.Errorf("scan unit_guid: %w", err)
		}
		guids = append(guids, g)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating unit_guids: %w", err)
	}

	return guids, nil
}

// scanMessages scans all rows into a slice of Message domain models.
func scanMessages(rows pgx.Rows) ([]domain.Message, error) {
	var messages []domain.Message

	for rows.Next() {
		var m domain.Message
		err := rows.Scan(
			&m.ID, &m.Number, &m.MQTT, &m.InvID, &m.UnitGUID, &m.MessageID,
			&m.MessageText, &m.MessageContext, &m.MessageClass, &m.MessageLevel,
			&m.VariableZone, &m.VariableAddress, &m.UseAsBlockStart,
			&m.Type, &m.BitNumberInRegister, &m.InvertBit, &m.SourceFile, &m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating messages: %w", err)
	}

	return messages, nil
}
