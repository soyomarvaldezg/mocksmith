package render

import (
	"encoding/json"
	"io"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/soyomarvaldezg/mocksmith/internal/generator"
	"github.com/soyomarvaldezg/mocksmith/internal/schema"
)

func JSONL(faker *gofakeit.Faker, w io.Writer, rows int, columns []schema.Column, duplicateRate float64, nullRate float64, corruptRate float64, outlierRate float64) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	hasNested := false
	for _, col := range columns {
		if len(col.Path) > 1 || col.ArrayKey != "" {
			hasNested = true
			break
		}
	}

	return RenderRows(faker, w, rows, columns, RowOptions{
		DuplicateRate: duplicateRate,
		NullRate:      nullRate,
		CorruptRate:   corruptRate,
		OutlierRate:   outlierRate,
	}, func(_ io.Writer, i int, values []RowValue) error {
		if hasNested {
			doc := buildNested(columns, values, faker)
			return enc.Encode(doc)
		}

		m := make(map[string]any, len(columns))
		for j, rv := range values {
			if rv.IsNull {
				m[columns[j].Name] = nil
			} else {
				m[columns[j].Name] = rv.Value
			}
		}
		return enc.Encode(m)
	})
}

func buildNested(columns []schema.Column, values []RowValue, faker *gofakeit.Faker) map[string]any {
	root := make(map[string]any)

	// Group array columns by their array key: map[arrayKey]map[fieldName]generator
	arrays := make(map[string]map[string]string)

	for j, col := range columns {
		rv := values[j]

		if col.ArrayKey != "" {
			if arrays[col.ArrayKey] == nil {
				arrays[col.ArrayKey] = make(map[string]string)
			}
			fieldName := col.Path[0]
			arrays[col.ArrayKey][fieldName] = col.Generator
		} else if len(col.Path) == 1 {
			if rv.IsNull {
				root[col.Path[0]] = nil
			} else {
				root[col.Path[0]] = rv.Value
			}
		} else {
			setNested(root, col.Path, rv)
		}
	}

	// Build arrays: generate 1-5 items per array
	for arrayKey, fields := range arrays {
		n := faker.IntN(5) + 1
		arr := make([]map[string]any, n)
		for idx := range n {
			item := make(map[string]any)
			for fieldName, gen := range fields {
				val, err := generator.Generate(faker, gen)
				if err != nil {
					item[fieldName] = nil
				} else {
					item[fieldName] = val
				}
			}
			arr[idx] = item
		}
		root[arrayKey] = arr
	}

	return root
}

func setNested(m map[string]any, path []string, rv RowValue) {
	if len(path) == 0 {
		return
	}

	if len(path) == 1 {
		if rv.IsNull {
			m[path[0]] = nil
		} else {
			m[path[0]] = rv.Value
		}
		return
	}

	if _, ok := m[path[0]]; !ok {
		m[path[0]] = make(map[string]any)
	}

	inner, ok := m[path[0]].(map[string]any)
	if !ok {
		return
	}

	setNested(inner, path[1:], rv)
}
