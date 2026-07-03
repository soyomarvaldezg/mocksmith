package render

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"codeberg.org/soyomarvaldezg/mocksmith/internal/schema"
	"github.com/brianvoe/gofakeit/v7"
)

var testColumns = []schema.Column{
	{Name: "id", Generator: "uuid"},
	{Name: "name", Generator: "firstname"},
	{Name: "email", Generator: "email"},
}

// ---- SQL tests ----

func TestSQLBasic(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := SQL(faker, &buf, "users", 3, testColumns, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("SQL: %v", err)
	}

	output := buf.String()
	if !strings.HasPrefix(output, "INSERT INTO users (id,name,email) VALUES") {
		t.Errorf("missing INSERT header")
	}
	if !strings.HasSuffix(output, ";\n") {
		t.Errorf("missing terminating semicolon, got: %q", output[len(output)-3:])
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 { // header + 3 rows
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}

func TestSQLReproducible(t *testing.T) {
	var b1, b2 bytes.Buffer

	SQL(gofakeit.New(42), &b1, "t", 3, testColumns, 0, 0, 0, 0)
	SQL(gofakeit.New(42), &b2, "t", 3, testColumns, 0, 0, 0, 0)

	if b1.String() != b2.String() {
		t.Errorf("SQL output not reproducible with same seed")
	}
}

func TestSQLDifferentSeed(t *testing.T) {
	var b1, b2 bytes.Buffer

	SQL(gofakeit.New(1), &b1, "t", 3, testColumns, 0, 0, 0, 0)
	SQL(gofakeit.New(9999), &b2, "t", 3, testColumns, 0, 0, 0, 0)

	if b1.String() == b2.String() {
		t.Error("different seeds produced identical output — extremely unlikely")
	}
}

func TestSQLNulls(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := SQL(faker, &buf, "t", 20, testColumns, 0, 0.5, 0, 0)
	if err != nil {
		t.Fatalf("SQL with nulls: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "NULL") {
		t.Error("no NULL values found with 50% null rate over 20 rows")
	}
}

func TestSQLNullRateZero(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := SQL(faker, &buf, "t", 10, testColumns, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("SQL: %v", err)
	}

	if strings.Contains(buf.String(), "NULL") {
		t.Error("NULL found with zero null rate")
	}
}

func TestSQLDuplicate(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := SQL(faker, &buf, "t", 30, testColumns, 0.5, 0, 0, 0)
	if err != nil {
		t.Fatalf("SQL with duplicates: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	// With 30 rows and 50% dup rate, some lines should repeat
	seen := make(map[string]int)
	for _, line := range lines[1:] { // skip header
		seen[line]++
	}

	hasDup := false
	for _, count := range seen {
		if count > 1 {
			hasDup = true
			break
		}
	}
	if !hasDup {
		t.Error("no duplicate rows found with 50% duplicate rate over 30 rows — unlikely")
	}
}

func TestSQLCorruption(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := SQL(faker, &buf, "t", 20, testColumns, 0, 0, 0.5, 0)
	if err != nil {
		t.Fatalf("SQL with corruption: %v", err)
	}

	output := buf.String()
	// Should have some non-alphanumeric from corruption
	hasGarbage := strings.ContainsAny(output, "!@#$%^&")
	if !hasGarbage {
		t.Error("no corruption characters found with 50% corrupt rate over 20 rows — unlikely")
	}
}

func TestSQLOutliers(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := SQL(faker, &buf, "t", 30, testColumns, 0, 0, 0, 0.5)
	if err != nil {
		t.Fatalf("SQL with outliers: %v", err)
	}

	output := buf.String()
	hasOutlier := strings.Contains(output, "XXXXXXXXXXXXXXXXXX") ||
		strings.Contains(output, "OR 1=1") ||
		strings.Contains(output, "ó🔥")
	if !hasOutlier {
		t.Error("no outlier patterns found with 50% outlier rate over 30 rows — unlikely")
	}
}

func TestSQLCombined(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := SQL(faker, &buf, "t", 50, testColumns, 0.2, 0.2, 0.2, 0.2)
	if err != nil {
		t.Fatalf("SQL combined: %v", err)
	}

	output := buf.String()
	if !strings.HasPrefix(output, "INSERT INTO") {
		t.Error("missing INSERT")
	}
	if !strings.HasSuffix(output, ";\n") {
		t.Error("missing semicolon")
	}
	if len(output) == 0 {
		t.Error("empty output")
	}
}

func TestSQLSingleRow(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := SQL(faker, &buf, "t", 1, testColumns, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("SQL single row: %v", err)
	}

	output := buf.String()
	if !strings.HasSuffix(output, ";\n") {
		t.Errorf("single row should end with semicolon, got: %q", output)
	}
	// Single row should NOT have a comma before the semicolon
	if strings.Contains(output, "),\n") {
		t.Error("single row should not have trailing comma")
	}
}

// ---- CSV tests ----

func TestCSVBasic(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := CSV(faker, &buf, 3, testColumns, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("CSV: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("CSV parse: %v", err)
	}

	if len(records) != 4 { // header + 3 rows
		t.Errorf("expected 4 records, got %d", len(records))
	}

	header := records[0]
	if header[0] != "id" || header[1] != "name" || header[2] != "email" {
		t.Errorf("unexpected header: %v", header)
	}

	for i, row := range records[1:] {
		if len(row) != 3 {
			t.Errorf("row %d: expected 3 fields, got %d", i, len(row))
		}
		for j, val := range row {
			if val == "" {
				t.Errorf("row %d col %d: empty value without null rate", i, j)
			}
		}
	}
}

func TestCSVReproducible(t *testing.T) {
	var b1, b2 bytes.Buffer

	CSV(gofakeit.New(42), &b1, 3, testColumns, 0, 0, 0, 0)
	CSV(gofakeit.New(42), &b2, 3, testColumns, 0, 0, 0, 0)

	if b1.String() != b2.String() {
		t.Errorf("CSV output not reproducible")
	}
}

func TestCSVNulls(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := CSV(faker, &buf, 30, testColumns, 0, 0.5, 0, 0)
	if err != nil {
		t.Fatalf("CSV with nulls: %v", err)
	}

	r := csv.NewReader(&buf)
	records, _ := r.ReadAll()

	emptyCount := 0
	for _, row := range records[1:] {
		for _, val := range row {
			if val == "" {
				emptyCount++
			}
		}
	}

	if emptyCount == 0 {
		t.Error("no empty CSV fields with 50% null rate over 30 rows — unlikely")
	}
}

func TestCSVCorruption(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := CSV(faker, &buf, 30, testColumns, 0, 0, 0.5, 0)
	if err != nil {
		t.Fatalf("CSV corrupt: %v", err)
	}

	output := buf.String()
	if !strings.ContainsAny(output, "!@#$%^&") {
		t.Error("no corruption chars in CSV with 50% corrupt rate — unlikely")
	}
}

// ---- JSONL tests ----

func TestJSONLBasic(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := JSONL(faker, &buf, 3, testColumns, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("JSONL: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Errorf("row %d: invalid JSON: %v", i, err)
		}
		for _, key := range []string{"id", "name", "email"} {
			if _, ok := m[key]; !ok {
				t.Errorf("row %d: missing key %q", i, key)
			}
		}
	}
}

func TestJSONLReproducible(t *testing.T) {
	var b1, b2 bytes.Buffer

	JSONL(gofakeit.New(42), &b1, 3, testColumns, 0, 0, 0, 0)
	JSONL(gofakeit.New(42), &b2, 3, testColumns, 0, 0, 0, 0)

	if b1.String() != b2.String() {
		t.Errorf("JSONL not reproducible")
	}
}

func TestJSONLNulls(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := JSONL(faker, &buf, 30, testColumns, 0, 0.5, 0, 0)
	if err != nil {
		t.Fatalf("JSONL nulls: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	nullCount := 0

	for _, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		for _, v := range m {
			if v == nil {
				nullCount++
			}
		}
	}

	if nullCount == 0 {
		t.Error("no JSON nulls with 50% null rate over 30 rows — unlikely")
	}
}

func TestJSONLNoHTML(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	err := JSONL(faker, &buf, 20, testColumns, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("JSONL: %v", err)
	}

	// With SetEscapeHTML(false), ampersands should not be escaped
	output := buf.String()
	if strings.Contains(output, "\\u0026") {
		t.Error("HTML escaping found despite SetEscapeHTML(false)")
	}
}

// ---- JSONL nested tests ----

func TestJSONLNestedObjects(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	columns := []schema.Column{
		{Name: "_id", Generator: "uuid", Path: []string{"_id"}},
		{Name: "name.first", Generator: "firstname", Path: []string{"name", "first"}},
		{Name: "name.last", Generator: "lastname", Path: []string{"name", "last"}},
		{Name: "email", Generator: "email", Path: []string{"email"}},
		{Name: "address.city", Generator: "city", Path: []string{"address", "city"}},
	}

	err := JSONL(faker, &buf, 3, columns, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("JSONL nested: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("row %d: invalid JSON: %v", i, err)
		}

		if _, ok := m["_id"]; !ok {
			t.Errorf("row %d: missing _id", i)
		}

		name, ok := m["name"].(map[string]any)
		if !ok {
			t.Errorf("row %d: name is not a nested object", i)
			continue
		}
		if _, ok := name["first"]; !ok {
			t.Errorf("row %d: missing name.first", i)
		}
		if _, ok := name["last"]; !ok {
			t.Errorf("row %d: missing name.last", i)
		}

		if _, ok := m["address"].(map[string]any); !ok {
			t.Errorf("row %d: address is not a nested object", i)
		}
	}
}

func TestJSONLArrays(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	columns := []schema.Column{
		{Name: "_id", Generator: "uuid", Path: []string{"_id"}},
		{Name: "orders[].id", Generator: "uuid", Path: []string{"id"}, ArrayKey: "orders"},
		{Name: "orders[].amount", Generator: "price", Path: []string{"amount"}, ArrayKey: "orders"},
	}

	err := JSONL(faker, &buf, 3, columns, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("JSONL arrays: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("row %d: invalid JSON: %v", i, err)
		}

		arr, ok := m["orders"].([]any)
		if !ok {
			t.Errorf("row %d: orders is not an array", i)
			continue
		}
		if len(arr) == 0 {
			t.Errorf("row %d: orders array is empty", i)
		}
		if len(arr) > 5 {
			t.Errorf("row %d: orders array has %d items (max 5)", i, len(arr))
		}

		for j, item := range arr {
			obj, ok := item.(map[string]any)
			if !ok {
				t.Errorf("row %d item %d: not an object", i, j)
				continue
			}
			if _, ok := obj["id"]; !ok {
				t.Errorf("row %d item %d: missing id", i, j)
			}
			if _, ok := obj["amount"]; !ok {
				t.Errorf("row %d item %d: missing amount", i, j)
			}
		}
	}
}

func TestJSONLMixedNestedAndFlat(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	columns := []schema.Column{
		{Name: "_id", Generator: "uuid", Path: []string{"_id"}},
		{Name: "email", Generator: "email", Path: []string{"email"}},
		{Name: "profile.bio", Generator: "sentence:10", Path: []string{"profile", "bio"}},
	}

	err := JSONL(faker, &buf, 2, columns, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("JSONL mixed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("row %d: %v", i, err)
		}
		if _, ok := m["_id"]; !ok {
			t.Errorf("row %d: missing _id", i)
		}
		if _, ok := m["email"]; !ok {
			t.Errorf("row %d: missing email", i)
		}
		profile, ok := m["profile"].(map[string]any)
		if !ok {
			t.Errorf("row %d: profile not nested", i)
			continue
		}
		if _, ok := profile["bio"]; !ok {
			t.Errorf("row %d: missing profile.bio", i)
		}
	}
}

func TestJSONLNestedReproducible(t *testing.T) {
	columns := []schema.Column{
		{Name: "_id", Generator: "uuid", Path: []string{"_id"}},
		{Name: "name.first", Generator: "firstname", Path: []string{"name", "first"}},
	}

	var b1, b2 bytes.Buffer
	JSONL(gofakeit.New(42), &b1, 3, columns, 0, 0, 0, 0)
	JSONL(gofakeit.New(42), &b2, 3, columns, 0, 0, 0, 0)

	if b1.String() != b2.String() {
		t.Errorf("nested JSONL not reproducible")
	}
}


func TestRenderRowsZeroRows(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	called := false
	err := RenderRows(faker, &buf, 0, testColumns, RowOptions{}, func(w io.Writer, i int, values []RowValue) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("RenderRows 0 rows: %v", err)
	}
	if called {
		t.Error("rowFn called for zero rows")
	}
}

func TestRenderRowsEmptyColumns(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	count := 0
	err := RenderRows(faker, &buf, 3, []schema.Column{}, RowOptions{}, func(w io.Writer, i int, values []RowValue) error {
		if len(values) != 0 {
			t.Errorf("expected 0 values, got %d", len(values))
		}
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("RenderRows empty columns: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 calls, got %d", count)
	}
}

func TestRenderRowsRowCount(t *testing.T) {
	faker := gofakeit.New(42)
	var buf bytes.Buffer

	count := 0
	err := RenderRows(faker, &buf, 7, testColumns, RowOptions{}, func(w io.Writer, i int, values []RowValue) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("RenderRows: %v", err)
	}
	if count != 7 {
		t.Errorf("expected 7 rows, got %d", count)
	}
}
