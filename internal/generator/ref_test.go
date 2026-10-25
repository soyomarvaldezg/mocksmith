package generator

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/soyomarvaldezg/mocksmith/internal/schema"
)

func TestGenerateRefInvalidFormat(t *testing.T) {
	faker := gofakeit.New(42)
	_, err := Generate(faker, "ref:no_dot")
	if err == nil {
		t.Error("expected error for ref without dot")
	}
}

func TestGenerateRefFileNotFound(t *testing.T) {
	faker := gofakeit.New(42)
	_, err := Generate(faker, "ref:/tmp/mocksmith_no_existe_99999.csv.id")
	if err == nil {
		t.Error("expected error for missing ref file")
	}
}

func TestGenerateRefBadColumn(t *testing.T) {
	f, _ := os.CreateTemp("", "mocksmith_ref_test_*.csv")
	defer os.Remove(f.Name())
	fmt.Fprintf(f, "id,name\nabc,John\n")
	f.Close()

	faker := gofakeit.New(42)
	_, err := Generate(faker, "ref:"+f.Name()+".badcol")
	if err == nil {
		t.Error("expected error for bad column")
	}
}

func TestGenerateRefCSV(t *testing.T) {
	f, _ := os.CreateTemp("", "mocksmith_ref_test_*.csv")
	defer os.Remove(f.Name())
	fmt.Fprintf(f, "id,name\nabc,John\ndef,Maria\nghi,Bob\n")
	f.Close()

	refCache = sync.Map{}

	faker := gofakeit.New(42)

	seen := make(map[string]bool)
	for i := 0; i < 20; i++ {
		val, err := Generate(faker, "ref:"+f.Name()+".id")
		if err != nil {
			t.Fatalf("ref CSV: %v", err)
		}
		seen[val] = true
	}

	expected := []string{"abc", "def", "ghi"}
	for _, e := range expected {
		if !seen[e] {
			t.Errorf("expected to see %q from CSV ref", e)
		}
	}
}

func TestGenerateRefJSONL(t *testing.T) {
	f, _ := os.CreateTemp("", "mocksmith_ref_test_*.jsonl")
	defer os.Remove(f.Name())
	fmt.Fprintf(f, `{"id":"a1","name":"X"}`+"\n")
	fmt.Fprintf(f, `{"id":"b2","name":"Y"}`+"\n")
	fmt.Fprintf(f, `{"id":"c3","name":"Z"}`+"\n")
	f.Close()

	refCache = sync.Map{}
	faker := gofakeit.New(42)

	val, err := Generate(faker, "ref:"+f.Name()+".id")
	if err != nil {
		t.Fatalf("ref JSONL: %v", err)
	}
	if val != "a1" && val != "b2" && val != "c3" {
		t.Errorf("unexpected ref value from JSONL: %q", val)
	}
}

func TestGenerateRefSQL(t *testing.T) {
	f, _ := os.CreateTemp("", "mocksmith_ref_test_*.sql")
	defer os.Remove(f.Name())
	fmt.Fprintf(f, `INSERT INTO t (id,name) VALUES`+"\n")
	fmt.Fprintf(f, `('x1','Alice'),`+"\n")
	fmt.Fprintf(f, `('x2','Bob'),`+"\n")
	fmt.Fprintf(f, `('x3','Carol');`+"\n")
	f.Close()

	refCache = sync.Map{}
	faker := gofakeit.New(42)

	val, err := Generate(faker, "ref:"+f.Name()+".id")
	if err != nil {
		t.Fatalf("ref SQL: %v", err)
	}
	if val != "x1" && val != "x2" && val != "x3" {
		t.Errorf("unexpected ref value from SQL: %q", val)
	}
}

func TestGenerateRefCached(t *testing.T) {
	f, _ := os.CreateTemp("", "mocksmith_ref_cache_*.csv")
	defer os.Remove(f.Name())
	fmt.Fprintf(f, "id\nv1\nv2\n")
	f.Close()

	refCache = sync.Map{}
	faker := gofakeit.New(42)

	for i := 0; i < 10; i++ {
		_, err := Generate(faker, "ref:"+f.Name()+".id")
		if err != nil {
			t.Fatalf("ref cached call %d: %v", i, err)
		}
	}

	if _, ok := refCache.Load(f.Name() + ".id"); !ok {
		t.Error("ref not cached after first call")
	}
}

func TestGenerateRefInRow(t *testing.T) {
	f, _ := os.CreateTemp("", "mocksmith_ref_row_*.csv")
	defer os.Remove(f.Name())
	fmt.Fprintf(f, "customer_id\nk1\nk2\nk3\n")
	f.Close()

	refCache = sync.Map{}
	faker := gofakeit.New(42)

	columns := []schema.Column{
		{Name: "id", Generator: "uuid"},
		{Name: "customer_ref", Generator: "ref:" + f.Name() + ".customer_id"},
	}

	row, err := GenerateRow(faker, columns)
	if err != nil {
		t.Fatalf("GenerateRow with ref: %v", err)
	}

	if row["customer_ref"] != "k1" && row["customer_ref"] != "k2" && row["customer_ref"] != "k3" {
		t.Errorf("unexpected ref in row: %q", row["customer_ref"])
	}

	if row["id"] == "" {
		t.Error("uuid column empty")
	}
}

func TestGenerateRefEmptyFile(t *testing.T) {
	f, _ := os.CreateTemp("", "mocksmith_ref_empty_*.csv")
	defer os.Remove(f.Name())
	fmt.Fprintf(f, "id\n")
	f.Close()

	refCache = sync.Map{}
	faker := gofakeit.New(42)

	_, err := Generate(faker, "ref:"+f.Name()+".id")
	if err == nil {
		t.Error("expected error for empty ref source (only header, no data rows)")
	}
}

func TestGenerateRefMultipleTypes(t *testing.T) {
	f1, _ := os.CreateTemp("", "mocksmith_ref_csv_*.csv")
	defer os.Remove(f1.Name())
	fmt.Fprintf(f1, "sku\nS1\nS2\n")
	f1.Close()

	f2, _ := os.CreateTemp("", "mocksmith_ref_jsonl_*.jsonl")
	defer os.Remove(f2.Name())
	fmt.Fprintf(f2, `{"code":"C1"}`+"\n")
	f2.Close()

	refCache = sync.Map{}
	faker := gofakeit.New(42)

	v1, err := Generate(faker, "ref:"+f1.Name()+".sku")
	if err != nil {
		t.Fatalf("ref CSV: %v", err)
	}
	v2, err := Generate(faker, "ref:"+f2.Name()+".code")
	if err != nil {
		t.Fatalf("ref JSONL: %v", err)
	}

	if v1 != "S1" && v1 != "S2" {
		t.Errorf("bad CSV ref: %q", v1)
	}
	if v2 != "C1" {
		t.Errorf("bad JSONL ref: %q", v2)
	}
}
