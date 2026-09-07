package services

import (
	"strconv"
	"strings"

	"paperviz/internal/models"
)

// ExtractTableData parses table-formatted text and returns structured evidence items.
func ExtractTableData(text string, pageNumber int) []models.NumericEvidence {
	tables := detectTables(text)
	var all []models.NumericEvidence
	for _, tbl := range tables {
		items := parseTable(tbl, pageNumber)
		all = append(all, items...)
	}
	return all
}

// detectTables finds contiguous table blocks (pipe or tab separated).
func detectTables(text string) []string {
	lines := strings.Split(text, "\n")
	var tables []string
	var buf []string
	inTable := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isTableRow(trimmed) {
			inTable = true
			buf = append(buf, trimmed)
		} else if inTable {
			tables = append(tables, strings.Join(buf, "\n"))
			buf = nil
			inTable = false
		}
	}
	if inTable && len(buf) > 0 {
		tables = append(tables, strings.Join(buf, "\n"))
	}
	return tables
}

// isTableRow returns true if line looks like a table row (pipe or multi-tab).
func isTableRow(line string) bool {
	if strings.Contains(line, "|") && strings.Count(line, "|") >= 2 {
		return true
	}
	// Tab-separated: at least 2 tabs
	return strings.Count(line, "\t") >= 2
}

// isSeparatorLine returns true if line is a table separator (---, ===, etc.).
func isSeparatorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.ReplaceAll(trimmed, "-", "")
	trimmed = strings.ReplaceAll(trimmed, "=", "")
	trimmed = strings.ReplaceAll(trimmed, "|", "")
	trimmed = strings.ReplaceAll(trimmed, " ", "")
	return len(trimmed) == 0 && len(strings.TrimSpace(line)) > 0
}

// parseTable extracts NumericEvidence from a single table block.
func parseTable(block string, pageNumber int) []models.NumericEvidence {
	lines := strings.Split(block, "\n")
	if len(lines) < 2 {
		return nil
	}

	// Find header row and separator
	headerIdx := -1
	sepIdx := -1
	for i, line := range lines {
		if isSeparatorLine(line) {
			if headerIdx < 0 {
				headerIdx = i - 1
			}
			sepIdx = i
			break
		}
	}
	if headerIdx < 0 || sepIdx < 0 {
		// No separator found — try first row as header
		headerIdx = 0
		sepIdx = -1
	}

	headers := splitRow(lines[headerIdx])
	if len(headers) < 2 {
		return nil
	}

	// Parse units from headers (e.g., "Accuracy (%)" → metric="accuracy", unit="%")
	type colInfo struct {
		metric string
		unit   string
	}
	cols := make([]colInfo, len(headers))
	for i, h := range headers {
		metric, unit := parseHeader(h)
		cols[i] = colInfo{metric: metric, unit: unit}
	}

	var items []models.NumericEvidence
	// Process data rows (after separator)
	startRow := sepIdx + 1
	if sepIdx < 0 {
		startRow = 1
	}
	for _, line := range lines[startRow:] {
		if isSeparatorLine(line) {
			continue
		}
		cells := splitRow(line)
		if len(cells) < 2 {
			continue
		}
		entity := strings.TrimSpace(cells[0])
		if entity == "" {
			continue
		}
		// Map remaining cells to columns
		for col := 1; col < len(cells) && col < len(cols); col++ {
			valStr := strings.TrimSpace(cells[col])
			if valStr == "" {
				continue // skip missing cells
			}
			val, err := parseFloat(valStr)
			if err != nil {
				continue // skip non-numeric cells
			}
			items = append(items, models.NumericEvidence{
				Metric: cols[col].metric,
				Entity: entity,
				Value:  val,
				Unit:   cols[col].unit,
				Source: models.EvidenceSource{Page: pageNumber, Text: block},
			})
		}
	}
	return items
}

// splitRow splits a table row by pipe or tab, preserving empty cell positions.
func splitRow(line string) []string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "|") {
		line = line[1:]
	}
	if strings.HasSuffix(line, "|") {
		line = line[:len(line)-1]
	}
	if strings.Contains(line, "|") {
		return splitPreserve(line, "|")
	}
	return splitPreserve(line, "\t")
}

// splitPreserve splits s by sep, trims each part, preserves empty cells.
func splitPreserve(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = strings.TrimSpace(p)
	}
	return out
}

// parseHeader extracts metric name and unit from a header cell like "Accuracy (%)".
func parseHeader(h string) (metric, unit string) {
	h = strings.TrimSpace(h)
	// Extract unit from parentheses: "Accuracy (%)" → unit="%"
	if idx := strings.LastIndex(h, "("); idx >= 0 {
		if end := strings.LastIndex(h, ")"); end > idx {
			unit = strings.TrimSpace(h[idx+1 : end])
			h = strings.TrimSpace(h[:idx])
		}
	}
	metric = strings.ToLower(h)
	return metric, unit
}

// parseFloat parses a number string, handling commas.
func parseFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}
