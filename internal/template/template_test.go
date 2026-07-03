package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseData(t *testing.T) {
	yaml := []byte(`
name: test
description: test template
seed: 42
tables:
  - name: users
    rows: 10
    format: sql
    columns:
      - name: id
        generator: uuid
      - name: name
        generator: firstname
`)

	tmpl, err := ParseData(yaml)
	if err != nil {
		t.Fatalf("ParseData: %v", err)
	}
	if tmpl.Name != "test" {
		t.Errorf("name = %q, want test", tmpl.Name)
	}
	if len(tmpl.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(tmpl.Tables))
	}
	if tmpl.Tables[0].Name != "users" {
		t.Errorf("table name = %q, want users", tmpl.Tables[0].Name)
	}
	if len(tmpl.Tables[0].Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(tmpl.Tables[0].Columns))
	}
}

func TestParseOverrides(t *testing.T) {
	overrides := ParseOverrides("customers:1000,orders:5000")

	if overrides["customers"] != 1000 {
		t.Errorf("customers = %d, want 1000", overrides["customers"])
	}
	if overrides["orders"] != 5000 {
		t.Errorf("orders = %d, want 5000", overrides["orders"])
	}
}

func TestParseOverridesEmpty(t *testing.T) {
	if ParseOverrides("") != nil {
		t.Error("empty overrides should be nil")
	}
}

func TestParseOverridesInvalid(t *testing.T) {
	overrides := ParseOverrides("bad,also=bad,ok:5")
	if overrides["ok"] != 5 {
		t.Errorf("ok = %d, want 5", overrides["ok"])
	}
	if len(overrides) != 1 {
		t.Errorf("expected 1 valid override, got %d", len(overrides))
	}
}

func TestGenerateCreatesFiles(t *testing.T) {
	dir := t.TempDir()

	yaml := []byte(`
name: testgen
seed: 42
tables:
  - name: items
    rows: 5
    format: csv
    columns:
      - name: id
        generator: uuid
      - name: label
        generator: word
`)

	tmpl, err := ParseData(yaml)
	if err != nil {
		t.Fatalf("ParseData: %v", err)
	}

	err = tmpl.Generate(GenerateOptions{OutputDir: dir})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	path := filepath.Join(dir, "items.csv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	if len(content) == 0 {
		t.Error("generated CSV is empty")
	}
}

func TestGenerateWithOverrides(t *testing.T) {
	dir := t.TempDir()

	yaml := []byte(`
name: test_override
seed: 42
tables:
  - name: data
    rows: 10
    format: jsonl
    columns:
      - name: id
        generator: uuid
`)

	tmpl, _ := ParseData(yaml)
	err := tmpl.Generate(GenerateOptions{
		OutputDir: dir,
		Overrides: map[string]int{"data": 3},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "data.jsonl"))
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines != 3 {
		t.Errorf("expected 3 lines (overridden from 10), got %d", lines)
	}
}

func TestGenerateMultipleFormats(t *testing.T) {
	dir := t.TempDir()

	yaml := []byte(`
name: multi
seed: 42
tables:
  - name: t1
    rows: 2
    format: sql
    columns:
      - name: id
        generator: uuid
  - name: t2
    rows: 2
    format: jsonl
    columns:
      - name: id
        generator: uuid
`)

	tmpl, _ := ParseData(yaml)
	err := tmpl.Generate(GenerateOptions{OutputDir: dir})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	for _, f := range []string{"t1.sql", "t2.jsonl"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing file: %s", f)
		}
	}
}
