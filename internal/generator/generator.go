package generator

import (
	"fmt"
	"strings"

	"codeberg.org/soyomarvaldezg/mocksmith/internal/schema"
	"github.com/brianvoe/gofakeit/v7"
)

func Generate(faker *gofakeit.Faker, kind string) (string, error) {
	if kind, ok := strings.CutPrefix(kind, "ref:"); ok {
		return generateRef(faker, kind)
	}

	val, err := faker.Generate("{" + kind + "}")
	if err != nil {
		return "", fmt.Errorf("generator %q: %w", kind, err)
	}
	if strings.HasPrefix(val, "{") && strings.HasSuffix(val, "}") {
		return "", fmt.Errorf("unknown generator: %s", kind)
	}
	return val, nil
}

func generateRef(faker *gofakeit.Faker, ref string) (string, error) {
	rs, err := loadRef(ref)
	if err != nil {
		return "", fmt.Errorf("ref %q: %w", ref, err)
	}
	return pickRef(faker, rs), nil
}

func GenerateRow(faker *gofakeit.Faker, columns []schema.Column) (map[string]string, error) {
	row := make(map[string]string)

	for _, col := range columns {
		value, err := Generate(faker, col.Generator)
		if err != nil {
			return nil, err
		}

		row[col.Name] = value
	}

	return row, nil
}
