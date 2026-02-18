package parser

import (
	"bufio"
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

	return parseTSVReader(f, fileName)
}

func parseTSVReader(r io.Reader, fileName string) ([]domain.Message, error) {
	scanner := bufio.NewScanner(r)

	// Skip the first two lines: comment line and header row.
	for i, label := range []string{"comment", "header"} {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("read tsv %s (line %d): %w", label, i+1, err)
			}
			return nil, fmt.Errorf("read tsv %s (line %d): unexpected end of file", label, i+1)
		}
	}

	var (
		messages []domain.Message
		firstErr *domain.ParseError
		line     = 2 // data starts at line 3
	)

	for scanner.Scan() {
		line++

		raw := scanner.Text()

		// Skip completely blank lines.
		if strings.TrimSpace(raw) == "" {
			continue
		}

		// Split by tab without any trimming — preserves empty fields.
		// A row "a\t\tb" correctly produces ["a", "", "b"].
		fields := strings.Split(raw, "\t")
		for i, v := range fields {
			fields[i] = strings.TrimSpace(v)
		}

		// Pad with empty strings if the row is shorter than expected.
		// This tolerates trailing empty columns that some editors omit.
		for len(fields) < totalColumns {
			fields = append(fields, "")
		}

		msg, parseErr := recordToMessage(fields, fileName, line)
		if parseErr != nil {
			if firstErr == nil {
				firstErr = parseErr
			}
			continue
		}

		messages = append(messages, *msg)
	}

	if err := scanner.Err(); err != nil {
		return messages, fmt.Errorf("scanning tsv file: %w", err)
	}

	if firstErr != nil {
		return messages, firstErr
	}

	return messages, nil
}

// recordToMessage maps a single TSV record to a domain.Message.
func recordToMessage(record []string, fileName string, line int) (*domain.Message, *domain.ParseError) {
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
		// Empty bit number is allowed — treat as 0.
		if strings.TrimSpace(record[colBitNumberInRegister]) == "" {
			bitNumber = 0
		} else {
			return nil, &domain.ParseError{
				File:    fileName,
				Line:    line,
				Column:  "bit_number_in_register",
				Message: fmt.Sprintf("invalid integer: %q", record[colBitNumberInRegister]),
			}
		}
	}

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
		UseAsBlockStart:     parseBool(record[colUseAsBlockStart]),
		Type:                strings.TrimSpace(record[colType]),
		BitNumberInRegister: bitNumber,
		InvertBit:           parseBool(record[colInvertBit]),
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
