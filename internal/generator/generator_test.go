package generator

import (
	"testing"

	"codeberg.org/soyomarvaldezg/mocksmith/internal/schema"
	"github.com/brianvoe/gofakeit/v7"
)

func TestGenerate(t *testing.T) {
	faker := gofakeit.New(42)

	tests := []struct {
		kind    string
		wantErr bool
	}{
		{"uuid", false},
		{"firstname", false},
		{"lastname", false},
		{"email", false},
		{"date", false},
		{"phone", false},
		{"city", false},
		{"country", false},
		{"ipv4address", false},
		{"price", false},
		{"creditcardnumber", false},
		{"sentence:30", false},
		{"number:1,100", false},
		{"no_existe_esto", true},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			val, err := Generate(faker, tt.kind)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate(%q) error = %v, wantErr %v", tt.kind, err, tt.wantErr)
			}
			if !tt.wantErr && val == "" {
				t.Errorf("Generate(%q) returned empty string", tt.kind)
			}
		})
	}
}

func TestGenerateReproducible(t *testing.T) {
	f1 := gofakeit.New(42)
	f2 := gofakeit.New(42)

	for _, kind := range []string{"uuid", "firstname", "lastname", "email", "phone", "city"} {
		v1, _ := Generate(f1, kind)
		v2, _ := Generate(f2, kind)
		if v1 != v2 {
			t.Errorf("Generate(%q) not reproducible: %q vs %q", kind, v1, v2)
		}
	}
}

func TestGenerateDifferentSeeds(t *testing.T) {
	f1 := gofakeit.New(1)
	f2 := gofakeit.New(9999)

	allSame := true
	for _, kind := range []string{"uuid", "firstname", "lastname"} {
		v1, _ := Generate(f1, kind)
		v2, _ := Generate(f2, kind)
		if v1 == v2 {
			allSame = allSame && true
		} else {
			allSame = false
		}
	}
	if allSame {
		t.Error("different seeds produced identical output for all generators — extremely unlikely")
	}
}

func TestGenerateRow(t *testing.T) {
	faker := gofakeit.New(42)

	columns := []schema.Column{
		{Name: "id", Generator: "uuid"},
		{Name: "first", Generator: "firstname"},
		{Name: "last", Generator: "lastname"},
	}

	row, err := GenerateRow(faker, columns)
	if err != nil {
		t.Fatalf("GenerateRow: %v", err)
	}

	if len(row) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(row))
	}

	for _, col := range columns {
		if v, ok := row[col.Name]; !ok {
			t.Errorf("missing key %q", col.Name)
		} else if v == "" {
			t.Errorf("empty value for %q", col.Name)
		}
	}
}

func TestGenerateRowUnknownGenerator(t *testing.T) {
	faker := gofakeit.New(42)
	columns := []schema.Column{
		{Name: "id", Generator: "uuid"},
		{Name: "bad", Generator: "no_such_generator"},
	}

	_, err := GenerateRow(faker, columns)
	if err == nil {
		t.Error("expected error for unknown generator, got nil")
	}
}

func TestGenerateRowReproducible(t *testing.T) {
	columns := []schema.Column{
		{Name: "id", Generator: "uuid"},
		{Name: "name", Generator: "firstname"},
	}

	f1 := gofakeit.New(42)
	f2 := gofakeit.New(42)

	r1, _ := GenerateRow(f1, columns)
	r2, _ := GenerateRow(f2, columns)

	for k := range r1 {
		if r1[k] != r2[k] {
			t.Errorf("row not reproducible: %q: %q vs %q", k, r1[k], r2[k])
		}
	}
}
