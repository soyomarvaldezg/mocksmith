package generator

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/brianvoe/gofakeit/v7"
)

type refSource struct {
	mu     sync.Mutex
	values []string
}

var refCache sync.Map

func loadRef(pathColumn string) (*refSource, error) {
	if cached, ok := refCache.Load(pathColumn); ok {
		return cached.(*refSource), nil
	}

	lastDot := strings.LastIndex(pathColumn, ".")
	if lastDot == -1 {
		return nil, fmt.Errorf("invalid ref format: %s (expected file.column)", pathColumn)
	}

	file := pathColumn[:lastDot]
	column := pathColumn[lastDot+1:]

	var values []string
	var err error

	switch {
	case strings.HasSuffix(file, ".csv"):
		values, err = loadCSVColumn(file, column)
	case strings.HasSuffix(file, ".jsonl"):
		values, err = loadJSONLColumn(file, column)
	case strings.HasSuffix(file, ".sql"):
		values, err = loadSQLColumn(file, column)
	default:
		return nil, fmt.Errorf("unsupported ref format: %s (use .csv, .jsonl, or .sql)", file)
	}

	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("ref %s: column %q has no values in %s", pathColumn, column, file)
	}

	rs := &refSource{values: values}
	refCache.Store(pathColumn, rs)
	return rs, nil
}

func loadCSVColumn(file, column string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV: %s", file)
	}

	colIdx := -1
	for i, h := range records[0] {
		if h == column {
			colIdx = i
			break
		}
	}
	if colIdx == -1 {
		return nil, fmt.Errorf("column %q not found in %s (header: %v)", column, file, records[0])
	}

	values := make([]string, 0, len(records)-1)
	for _, row := range records[1:] {
		if colIdx < len(row) && row[colIdx] != "" {
			values = append(values, row[colIdx])
		}
	}
	return values, nil
}

func loadJSONLColumn(file, column string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var values []string
	dec := json.NewDecoder(f)
	for dec.More() {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			return nil, err
		}
		if v, ok := m[column]; ok && v != nil {
			values = append(values, fmt.Sprint(v))
		}
	}
	return values, nil
}

func loadSQLColumn(file, column string) ([]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	if len(lines) < 1 {
		return nil, fmt.Errorf("empty SQL file: %s", file)
	}

	headerLine := lines[0]
	colIdx := columnIndex(headerLine, column)
	if colIdx == -1 {
		return nil, fmt.Errorf("column %q not found in %s", column, file)
	}

	var values []string
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimSuffix(line, ";")
		line = strings.TrimSuffix(line, ",")
		vals := parseSQLRow(line)
		if colIdx < len(vals) && vals[colIdx] != "NULL" {
			values = append(values, vals[colIdx])
		}
	}
	return values, nil
}

func columnIndex(headerLine, column string) int {
	start := strings.Index(headerLine, "(")
	end := strings.Index(headerLine, ")")
	if start == -1 || end == -1 {
		return -1
	}
	cols := strings.Split(headerLine[start+1:end], ",")
	for i, c := range cols {
		if strings.TrimSpace(c) == column {
			return i
		}
	}
	return -1
}

func parseSQLRow(line string) []string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "(") {
		return nil
	}
	line = line[1:]
	line = strings.TrimSuffix(line, ")")

	var vals []string
	inQuote := false
	current := strings.Builder{}
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '\'' {
			if inQuote && i+1 < len(line) && line[i+1] == '\'' {
				current.WriteByte('\'')
				i++
			} else {
				inQuote = !inQuote
			}
		} else if ch == ',' && !inQuote {
			vals = append(vals, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	vals = append(vals, strings.TrimSpace(current.String()))
	return vals
}

func pickRef(faker *gofakeit.Faker, rs *refSource) string {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.values[faker.IntN(len(rs.values))]
}
