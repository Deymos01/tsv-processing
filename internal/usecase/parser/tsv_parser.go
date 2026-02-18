package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Deymos01/tsv-processing/internal/domain"
)

// column indices in the TSV file (0-based).
const (
	colNumber = iota
	colMQTT
	colInvID
	colUnitGUID
	colMessageID
	colMessageText
	colContext
	colMessageClass
	colMessageLevel
	colVariableZone
	colVariableAddress
	colUseAsBlockStart
	colType
	colBitNumberInRegister
	colInvertBit
	totalColumns = 15
)

// ParseTSV reads a .tsv file and returns parsed messages.
// If any rows fail to parse, a *domain.ParseError is returned alongside
// the successfully parsed rows — callers must handle both.
func ParseTSV(filePath, fileName string) ([]domain.Message, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open tsv file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.LazyQuotes = true // TSV files may contain unescaped quotes
	r.TrimLeadingSpace = true

	// Skip header row.
	if _, err = r.Read(); err != nil {
		return nil, fmt.Errorf("read tsv header: %w", err)
	}

	var (
		messages []domain.Message
		firstErr *domain.ParseError
		line     = 1 // header was line 1; data starts at line 2
	)

	for {
		line++
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if firstErr == nil {
				firstErr = &domain.ParseError{
					File:    fileName,
					Line:    line,
					Column:  "—",
					Message: fmt.Sprintf("csv read error: %s", err.Error()),
				}
			}
			continue
		}

		msg, parseErr := recordToMessage(record, fileName, line)
		if parseErr != nil {
			if firstErr == nil {
				firstErr = parseErr
			}
			continue
		}

		messages = append(messages, *msg)
	}

	if firstErr != nil {
		return messages, firstErr
	}

	return messages, nil
}

// recordToMessage maps a single TSV record to a domain.Message.
func recordToMessage(record []string, fileName string, line int) (*domain.Message, *domain.ParseError) {
	if len(record) < totalColumns {
		return nil, &domain.ParseError{
			File:    fileName,
			Line:    line,
			Column:  "—",
			Message: fmt.Sprintf("expected %d columns, got %d", totalColumns, len(record)),
		}
	}

	number, err := strconv.Atoi(strings.TrimSpace(record[colNumber]))
	if err != nil {
		return nil, &domain.ParseError{
			File:    fileName,
			Line:    line,
			Column:  "number",
			Message: fmt.Sprintf("invalid integer: %q", record[colNumber]),
		}
	}

	bitNumber, err := strconv.Atoi(strings.TrimSpace(record[colBitNumberInRegister]))
	if err != nil {
		return nil, &domain.ParseError{
			File:    fileName,
			Line:    line,
			Column:  "bit_number_in_register",
			Message: fmt.Sprintf("invalid integer: %q", record[colBitNumberInRegister]),
		}
	}

	useAsBlockStart := parseBool(record[colUseAsBlockStart])
	invertBit := parseBool(record[colInvertBit])

	msg := &domain.Message{
		Number:              number,
		MQTT:                strings.TrimSpace(record[colMQTT]),
		InvID:               strings.TrimSpace(record[colInvID]),
		UnitGUID:            strings.TrimSpace(record[colUnitGUID]),
		MessageID:           strings.TrimSpace(record[colMessageID]),
		MessageText:         strings.TrimSpace(record[colMessageText]),
		MessageContext:      strings.TrimSpace(record[colContext]),
		MessageClass:        strings.TrimSpace(record[colMessageClass]),
		MessageLevel:        strings.TrimSpace(record[colMessageLevel]),
		VariableZone:        strings.TrimSpace(record[colVariableZone]),
		VariableAddress:     strings.TrimSpace(record[colVariableAddress]),
		UseAsBlockStart:     useAsBlockStart,
		Type:                strings.TrimSpace(record[colType]),
		BitNumberInRegister: bitNumber,
		InvertBit:           invertBit,
		SourceFile:          fileName,
	}

	return msg, nil
}

// parseBool interprets common TSV boolean representations.
// Treats "1", "true", "yes" (case-insensitive) as true, everything else as false.
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "1" || s == "true" || s == "yes"
}
