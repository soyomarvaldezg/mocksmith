package render

import (
	"encoding/csv"
	"io"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/soyomarvaldezg/mocksmith/internal/schema"
)

func CSV(faker *gofakeit.Faker, w io.Writer, rows int, columns []schema.Column, duplicateRate float64, nullRate float64, corruptRate float64, outlierRate float64) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := make([]string, 0, len(columns))
	for _, col := range columns {
		header = append(header, col.Name)
	}
	cw.Write(header)

	return RenderRows(faker, w, rows, columns, RowOptions{
		DuplicateRate: duplicateRate,
		NullRate:      nullRate,
		CorruptRate:   corruptRate,
		OutlierRate:   outlierRate,
	}, func(_ io.Writer, i int, values []RowValue) error {
		row := make([]string, 0, len(values))
		for _, rv := range values {
			if rv.IsNull {
				row = append(row, "")
			} else {
				row = append(row, rv.Value)
			}
		}
		cw.Write(row)
		return nil
	})
}
