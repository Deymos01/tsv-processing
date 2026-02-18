package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Deymos01/tsv-processing/internal/domain"
)

// RTFGenerator implements usecase/generator.OutputGenerator using plain RTF markup.
type RTFGenerator struct {
	outputDir string
	log       *zap.Logger
}

// NewRTFGenerator creates a new RTFGenerator.
func NewRTFGenerator(outputDir string, log *zap.Logger) (*RTFGenerator, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	return &RTFGenerator{
		outputDir: outputDir,
		log:       log,
	}, nil
}

// GenerateReport creates an RTF file named <unitGUID>.rtf in the output directory.
// The file contains a formatted table with all messages for that unit_guid.
func (g *RTFGenerator) GenerateReport(unitGUID string, messages []domain.Message) error {
	if len(messages) == 0 {
		g.log.Warn("no messages to generate report for", zap.String("unit_guid", unitGUID))
		return nil
	}

	content := buildReportRTF(unitGUID, messages)
	return g.writeFile(unitGUID+".rtf", content)
}

// GenerateErrorReport creates an RTF file describing a parse failure.
// The file is named after the source file with an "_error" suffix.
func (g *RTFGenerator) GenerateErrorReport(sourceFile string, parseErr error) error {
	baseName := strings.TrimSuffix(sourceFile, filepath.Ext(sourceFile))
	content := buildErrorRTF(sourceFile, parseErr)
	return g.writeFile(baseName+"_error.rtf", content)
}

// writeFile writes content to outputDir/<name>, creating the file if it doesn't exist.
func (g *RTFGenerator) writeFile(name string, content []byte) error {
	path := filepath.Join(g.outputDir, name)

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write rtf file %q: %w", path, err)
	}

	g.log.Info("rtf file written", zap.String("path", path))
	return nil
}

// buildReportRTF constructs a complete RTF document for a unit_guid report.
func buildReportRTF(unitGUID string, messages []domain.Message) []byte {
	var b bytes.Buffer

	writeRTFHeader(&b)
	writeRTFTitle(&b, fmt.Sprintf("Report for Unit GUID: %s", unitGUID))
	writeRTFParagraph(&b, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	writeRTFParagraph(&b, fmt.Sprintf("Total messages: %d", len(messages)))
	writeRTFNewLine(&b)

	// Table header.
	headers := []string{
		"n", "MQTT", "InvID", "Unit GUID", "Message ID",
		"Message Text", "Context", "Class", "Level",
		"Variable Zone", "Variable Address", "Block Start",
		"Type", "Bit n", "Invert Bit",
	}
	writeRTFTableRow(&b, headers, true)

	// Table rows.
	for _, m := range messages {
		row := []string{
			strconv.Itoa(m.Number),
			m.MQTT,
			m.InvID,
			m.UnitGUID,
			m.MessageID,
			m.MessageText,
			m.MessageContext,
			m.MessageClass,
			m.MessageLevel,
			m.VariableZone,
			m.VariableAddress,
			boolToStr(m.UseAsBlockStart),
			m.Type,
			strconv.Itoa(m.BitNumberInRegister),
			boolToStr(m.InvertBit),
		}
		writeRTFTableRow(&b, row, false)
	}

	writeRTFFooter(&b)
	return b.Bytes()
}

// buildErrorRTF constructs an RTF document describing a parse error.
func buildErrorRTF(sourceFile string, parseErr error) []byte {
	var b bytes.Buffer

	writeRTFHeader(&b)
	writeRTFTitle(&b, "File Processing Error Report")
	writeRTFParagraph(&b, fmt.Sprintf("Source file: %s", sourceFile))
	writeRTFParagraph(&b, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	writeRTFNewLine(&b)
	writeRTFTitle(&b, "Error Details")
	writeRTFParagraph(&b, escapeRTF(parseErr.Error()))
	writeRTFFooter(&b)

	return b.Bytes()
}

// ---------------------------------------------------------------------------
// Low-level RTF helpers
// ---------------------------------------------------------------------------

// writeRTFHeader writes the RTF document preamble with font and colour tables.
func writeRTFHeader(b *bytes.Buffer) {
	b.WriteString(`{\rtf1\ansi\deff0`)
	b.WriteString(`{\fonttbl{\f0\froman\fcharset204 Times New Roman;}{\f1\fswiss\fcharset204 Arial;}}`)
	b.WriteString(`{\colortbl;\red0\green0\blue0;\red0\green0\blue128;\red200\green200\blue200;}`)
	b.WriteString(`\widowctrl\hyphauto`)
	b.WriteString("\n")
}

// writeRTFFooter closes the RTF document.
func writeRTFFooter(b *bytes.Buffer) {
	b.WriteString("}")
}

// writeRTFTitle writes a bold, large-font paragraph.
func writeRTFTitle(b *bytes.Buffer, text string) {
	b.WriteString(`\pard\sb240\sa120\b\f1\fs28 `)
	b.WriteString(escapeRTF(text))
	b.WriteString(`\b0\par`)
	b.WriteString("\n")
}

// writeRTFParagraph writes a standard body paragraph.
func writeRTFParagraph(b *bytes.Buffer, text string) {
	b.WriteString(`\pard\sb60\sa60\f0\fs22 `)
	b.WriteString(escapeRTF(text))
	b.WriteString(`\par`)
	b.WriteString("\n")
}

// writeRTFNewLine writes an empty paragraph (vertical spacing).
func writeRTFNewLine(b *bytes.Buffer) {
	b.WriteString(`\pard\par`)
	b.WriteString("\n")
}

// writeRTFTableRow writes a single row of a simulated RTF table.
// RTF tables use \trowd/\row markup; each cell is separated by \cell.
func writeRTFTableRow(b *bytes.Buffer, cells []string, isHeader bool) {
	const cellWidth = 1200 // twips (~2.1 cm per cell)

	b.WriteString(`\trowd\trgaph108\trleft-108`)

	// Define cell boundaries.
	for i := range cells {
		pos := cellWidth * (i + 1)
		b.WriteString(fmt.Sprintf(`\cellx%d`, pos))
	}

	b.WriteString("\n")

	// Write cell contents.
	for _, cell := range cells {
		if isHeader {
			b.WriteString(`\pard\intbl\b\f1\fs18 `)
		} else {
			b.WriteString(`\pard\intbl\b0\f0\fs18 `)
		}
		b.WriteString(escapeRTF(cell))
		b.WriteString(`\cell`)
		b.WriteString("\n")
	}

	b.WriteString(`\row`)
	b.WriteString("\n")
}

// escapeRTF escapes characters that have special meaning in RTF.
// Cyrillic characters are encoded as Unicode escape sequences (\uN?).
func escapeRTF(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		switch {
		case r == '\\':
			b.WriteString(`\\`)
		case r == '{':
			b.WriteString(`\{`)
		case r == '}':
			b.WriteString(`\}`)
		case r == '\n':
			b.WriteString(`\line `)
		case r == '\r':
			// skip carriage return
		case r > 127:
			// Encode non-ASCII (including Cyrillic) as RTF Unicode escape.
			b.WriteString(fmt.Sprintf(`\u%d?`, int32(r)))
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

// boolToStr converts a bool to a human-readable string.
func boolToStr(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}
