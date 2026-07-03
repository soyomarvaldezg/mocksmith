package generator

import (
	"fmt"
	"strconv"
	"strings"

	"codeberg.org/soyomarvaldezg/mocksmith/internal/schema"
	"github.com/brianvoe/gofakeit/v7"
)

func Generate(faker *gofakeit.Faker, kind string) (string, error) {
	if parts, ok := strings.CutPrefix(kind, "ref:"); ok {
		return generateRef(faker, parts)
	}

	if parts, ok := strings.CutPrefix(kind, "weighted:"); ok {
		return generateWeighted(faker, parts)
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

func generateWeighted(faker *gofakeit.Faker, spec string) (string, error) {
	parts := strings.Split(spec, ",")
	if len(parts) < 2 {
		return "", fmt.Errorf("weighted requires at least 2 items: weighted:value:weight,value:weight")
	}

	type pair struct {
		value  string
		weight int
	}

	items := make([]pair, 0, len(parts))
	total := 0

	for _, part := range parts {
		idx := strings.LastIndex(part, ":")
		if idx == -1 {
			return "", fmt.Errorf("weighted item %q: expected value:weight format", part)
		}
		value := part[:idx]
		weight, err := strconv.Atoi(part[idx+1:])
		if err != nil || weight < 0 {
			return "", fmt.Errorf("weighted item %q: weight must be a non-negative integer", part)
		}
		if weight > 0 {
			items = append(items, pair{value, weight})
			total += weight
		}
	}

	if total == 0 {
		return "", fmt.Errorf("weighted: total weight is 0")
	}

	roll := faker.Float64() * float64(total)
	cumulative := 0
	for _, item := range items {
		cumulative += item.weight
		if float64(cumulative) >= roll {
			return item.value, nil
		}
	}

	return items[len(items)-1].value, nil
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
