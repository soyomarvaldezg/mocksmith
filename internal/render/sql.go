package render

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"codeberg.org/soyomarvaldezg/mocksmith/internal/generator"
	"codeberg.org/soyomarvaldezg/mocksmith/internal/schema"
	"github.com/brianvoe/gofakeit/v7"
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

func RenderRows(faker *gofakeit.Faker, w io.Writer, rows int, columns []schema.Column, opts RowOptions, rowFn func(w io.Writer, i int, values []RowValue) error) error {
	rowCache := make([]map[string]string, 0, rows)

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
			rowCache = append(rowCache, row)
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
				parts = append(parts, sqlQuote(rv.Value))
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

func sqlQuote(val string) string {
	if isNumeric(val) {
		return val
	}
	escaped := strings.ReplaceAll(val, "'", "''")
	return "'" + escaped + "'"
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	dots := 0
	for i, c := range s {
		if c == '-' && i == 0 {
			continue
		}
		if c == '.' {
			dots++
			if dots > 1 {
				return false
			}
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
