package schema

import (
	"testing"
)

func TestParseColumns(t *testing.T) {
	columns, err := ParseColumns([]string{"id=uuid", "name=firstname", "email=email"})
	if err != nil {
		t.Fatalf("ParseColumns: %v", err)
	}

	if len(columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(columns))
	}

	expected := []Column{
		{Name: "id", Generator: "uuid"},
		{Name: "name", Generator: "firstname"},
		{Name: "email", Generator: "email"},
	}

	for i, col := range columns {
		if col.Name != expected[i].Name || col.Generator != expected[i].Generator {
			t.Errorf("column[%d] = %+v, want %+v", i, col, expected[i])
		}
	}
}

func TestParseColumnsInvalid(t *testing.T) {
	_, err := ParseColumns([]string{"badformat"})
	if err == nil {
		t.Error("expected error for missing '='")
	}
}

func TestParseColumnsEmpty(t *testing.T) {
	columns, err := ParseColumns([]string{})
	if err != nil {
		t.Fatalf("ParseColumns empty: %v", err)
	}
	if len(columns) != 0 {
		t.Errorf("expected 0 columns, got %d", len(columns))
	}
}

func TestParseColumnsValueWithEquals(t *testing.T) {
	columns, err := ParseColumns([]string{"name=firstname", "extra=key=value"})
	if err != nil {
		t.Fatalf("ParseColumns: %v", err)
	}
	if columns[1].Generator != "key=value" {
		t.Errorf("expected generator 'key=value', got %q", columns[1].Generator)
	}
}

func TestParseColumnsRejectsMalformedPaths(t *testing.T) {
	for _, def := range []string{
		"orders[]=uuid",
		"[]=uuid",
		"orders[].=uuid",
		"address..city=city",
		"=uuid",
	} {
		if _, err := ParseColumns([]string{def}); err == nil {
			t.Errorf("ParseColumns(%q) expected an error", def)
		}
	}
}
