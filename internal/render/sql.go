package render

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/soyomarvaldezg/mocksmith/internal/generator"
	"github.com/soyomarvaldezg/mocksmith/internal/schema"
)

type RowValue struct {
	Value  string
	IsNull bool
}

type RowOptions struct {
	DuplicateRate float64
	NullRate      float64
	CorruptRate   float64
	OutlierRate   float64
}

const maxDuplicateCache = 100_000

type SQLDialect string

const (
	SQLStandard SQLDialect = "standard"
	SQLMySQL    SQLDialect = "mysql"
)

func ParseSQLDialect(value string) (SQLDialect, error) {
	switch SQLDialect(value) {
	case "", SQLStandard:
		return SQLStandard, nil
	case SQLMySQL:
		return SQLMySQL, nil
	default:
		return "", fmt.Errorf("unsupported SQL dialect %q (use standard or mysql)", value)
	}
}

func RenderRows(faker *gofakeit.Faker, w io.Writer, rows int, columns []schema.Column, opts RowOptions, rowFn func(w io.Writer, i int, values []RowValue) error) error {
	cacheCapacity := 0
	if opts.DuplicateRate > 0 && rows > 0 {
		cacheCapacity = min(rows, maxDuplicateCache)
	}
	rowCache := make([]map[string]string, 0, cacheCapacity)

	for i := range rows {
		var row map[string]string

		if len(rowCache) > 0 && faker.Float64() < opts.DuplicateRate {
			row = rowCache[faker.IntN(len(rowCache))]
		} else {
			var err error
			row, err = generator.GenerateRow(faker, columns)
			if err != nil {
				return err
			}
			if opts.DuplicateRate > 0 && len(rowCache) < maxDuplicateCache {
				rowCache = append(rowCache, row)
			}
		}

		values := make([]RowValue, 0, len(columns))
		for _, col := range columns {
			rv := RowValue{Value: row[col.Name]}

			if faker.Float64() < opts.NullRate {
				rv.IsNull = true
				rv.Value = ""
				values = append(values, rv)
				continue
			}

			if faker.Float64() < opts.OutlierRate {
				rv.Value = generateOutlier(faker)
			} else if faker.Float64() < opts.CorruptRate {
				rv.Value = corrupt(faker, rv.Value)
			}
			values = append(values, rv)
		}

		if err := rowFn(w, i, values); err != nil {
			return err
		}
	}

	return nil
}

func SQL(faker *gofakeit.Faker, w io.Writer, table string, rows int, columns []schema.Column, duplicateRate float64, nullRate float64, corruptRate float64, outlierRate float64) error {
	return SQLWithDialect(faker, w, table, rows, columns, duplicateRate, nullRate, corruptRate, outlierRate, SQLStandard)
}

func SQLWithDialect(faker *gofakeit.Faker, w io.Writer, table string, rows int, columns []schema.Column, duplicateRate float64, nullRate float64, corruptRate float64, outlierRate float64, dialect SQLDialect) error {
	names := make([]string, 0, len(columns))
	for _, col := range columns {
		names = append(names, col.Name)
	}

	fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES\n", table, strings.Join(names, ","))

	return RenderRows(faker, w, rows, columns, RowOptions{
		DuplicateRate: duplicateRate,
		NullRate:      nullRate,
		CorruptRate:   corruptRate,
		OutlierRate:   outlierRate,
	}, func(w io.Writer, i int, values []RowValue) error {
		parts := make([]string, 0, len(values))
		for _, rv := range values {
			if rv.IsNull {
				parts = append(parts, "NULL")
			} else {
				parts = append(parts, sqlQuote(rv.Value, dialect))
			}
		}

		if i == rows-1 {
			fmt.Fprintf(w, "(%s);\n", strings.Join(parts, ","))
		} else {
			fmt.Fprintf(w, "(%s),\n", strings.Join(parts, ","))
		}
		return nil
	})
}

var garbage = []rune("!@#$%^&*()_+-=[]{}|;':\",./<>?`~ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789😀💩🔥")

func corrupt(faker *gofakeit.Faker, s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}

	mutations := max(1, faker.IntN(max(1, utf8.RuneCountInString(s)*4/5))+utf8.RuneCountInString(s)/5)
	for range mutations {
		pos := faker.IntN(len(runes))
		runes[pos] = garbage[faker.IntN(len(garbage))]
	}
	return string(runes)
}

func generateOutlier(faker *gofakeit.Faker) string {
	switch faker.IntN(5) {
	case 0:
		return ""
	case 1:
		return strings.Repeat("X", 500)
	case 2:
		return "' OR 1=1 --"
	case 3:
		return strings.Repeat("Z", 1024)
	default:
		return "ó🔥测试𝒜𝓈☃"
	}
}

func sqlQuote(val string, dialect SQLDialect) string {
	if isNumeric(val) {
		return val
	}
	if dialect == SQLMySQL {
		val = strings.ReplaceAll(val, `\`, `\\`)
		val = strings.ReplaceAll(val, "\x00", `\0`)
	}
	escaped := strings.ReplaceAll(val, "'", "''")
	return "'" + escaped + "'"
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}

	start := 0
	if s[0] == '-' {
		start = 1
		if len(s) == 1 {
			return false
		}
	}

	dots := 0
	digits := 0
	for i := start; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			digits++
		case c == '.':
			dots++
			if dots > 1 {
				return false
			}
		default:
			return false
		}
	}

	if digits == 0 {
		return false
	}

	// Reject leading-zero integers such as "007"; keep "0" and "0.5".
	unsigned := s[start:]
	if len(unsigned) > 1 && unsigned[0] == '0' && unsigned[1] != '.' {
		return false
	}

	return true
}
