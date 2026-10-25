package cmd

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func execute(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()

	buf := new(noopBuffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	return buf.out.String(), err
}

// noopBuffer collects stdout writes while still letting format commands
// write via their own io.Writer (openOutput path-based). We also capture
// cobra-level output like version and help text.
type noopBuffer struct {
	out strings.Builder
}

func (b *noopBuffer) Write(p []byte) (int, error) {
	b.out.Write(p)
	return len(p), nil
}

func resetFlags() {
	sqlSeed = 0
	sqlDuplicateRate = 0
	sqlNullRate = 0
	sqlCorruptRate = 0
	sqlOutlierRate = 0
	sqlDialect = "standard"
	sqlOutput = ""

	csvSeed = 0
	csvDuplicateRate = 0
	csvNullRate = 0
	csvCorruptRate = 0
	csvOutlierRate = 0
	csvOutput = ""

	jsonlSeed = 0
	jsonlDuplicateRate = 0
	jsonlNullRate = 0
	jsonlCorruptRate = 0
	jsonlOutlierRate = 0
	jsonlOutput = ""

	tmplOutput = ""
	tmplOverrides = ""
}

func TestRootCmd(t *testing.T) {
	output, err := execute(t)
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	if !strings.Contains(output, "mocksmith") {
		t.Errorf("expected usage to mention mocksmith")
	}
}

func TestVersionCmd(t *testing.T) {
	output, err := execute(t, "version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(output, "mocksmith") {
		t.Errorf("version output should contain mocksmith, got: %q", output)
	}
}

func TestSQLOutput(t *testing.T) {
	resetFlags()

	path := t.TempDir() + "/test.sql"
	sqlSeed = 42
	sqlOutput = path

	_, err := execute(t, "sql", "users", "3", "id=uuid", "name=firstname")
	if err != nil {
		t.Fatalf("sql: %v", err)
	}

	content := readFile(t, path)
	if !strings.Contains(content, "INSERT INTO users (id,name) VALUES") {
		t.Error("missing INSERT header")
	}
	if !strings.HasSuffix(strings.TrimSpace(content), ");") {
		t.Errorf("missing terminating semicolon, got: %q", content[len(content)-20:])
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) != 4 { // header + 3 rows
		t.Errorf("expected 4 lines, got %d: %s", len(lines), content)
	}
}

func TestSQLReproducible(t *testing.T) {
	dir := t.TempDir()

	resetFlags()
	sqlSeed = 42
	sqlOutput = dir + "/a.sql"
	execute(t, "sql", "t", "3", "id=uuid", "name=firstname")
	a := readFile(t, dir+"/a.sql")

	resetFlags()
	sqlSeed = 42
	sqlOutput = dir + "/b.sql"
	execute(t, "sql", "t", "3", "id=uuid", "name=firstname")
	b := readFile(t, dir+"/b.sql")

	if a != b {
		t.Error("output not reproducible with same seed")
	}
}

func TestSQLMissingArgs(t *testing.T) {
	resetFlags()
	_, err := execute(t, "sql", "users")
	if err == nil {
		t.Error("expected error for missing rows")
	}
}

func TestSQLInvalidRows(t *testing.T) {
	resetFlags()
	_, err := execute(t, "sql", "users", "abc", "id=uuid")
	if err == nil {
		t.Error("expected error for non-integer rows")
	}
}

func TestSQLBadColumn(t *testing.T) {
	resetFlags()
	_, err := execute(t, "sql", "users", "1", "badcolumn")
	if err == nil {
		t.Error("expected error for missing = in column")
	}
}

func TestSQLInvalidDialect(t *testing.T) {
	resetFlags()
	_, err := execute(t, "sql", "users", "1", "id=uuid", "--dialect", "oracle")
	if err == nil {
		t.Error("expected error for unsupported SQL dialect")
	}
}

func TestCSVOutput(t *testing.T) {
	resetFlags()

	path := t.TempDir() + "/test.csv"
	csvSeed = 42
	csvOutput = path

	_, err := execute(t, "csv", "3", "id=uuid", "name=firstname", "email=email")
	if err != nil {
		t.Fatalf("csv: %v", err)
	}

	content := readFile(t, path)
	if !strings.Contains(content, "id,name,email") {
		t.Error("missing CSV header")
	}
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}

func TestCSVReproducible(t *testing.T) {
	dir := t.TempDir()

	resetFlags()
	csvSeed = 42
	csvOutput = dir + "/a.csv"
	execute(t, "csv", "3", "id=uuid")
	a := readFile(t, dir+"/a.csv")

	resetFlags()
	csvSeed = 42
	csvOutput = dir + "/b.csv"
	execute(t, "csv", "3", "id=uuid")
	b := readFile(t, dir+"/b.csv")

	if a != b {
		t.Error("CSV not reproducible")
	}
}

func TestCSVMissingArgs(t *testing.T) {
	resetFlags()
	_, err := execute(t, "csv")
	if err == nil {
		t.Error("expected error for missing args")
	}
}

func TestJSONLOutput(t *testing.T) {
	resetFlags()

	path := t.TempDir() + "/test.jsonl"
	jsonlSeed = 42
	jsonlOutput = path

	_, err := execute(t, "jsonl", "3", "id=uuid", "name=firstname")
	if err != nil {
		t.Fatalf("jsonl: %v", err)
	}

	content := readFile(t, path)
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Errorf("line %d: invalid JSON: %v", i, err)
		}
		if _, ok := m["id"]; !ok {
			t.Errorf("line %d: missing id", i)
		}
	}
}

func TestJSONLReproducible(t *testing.T) {
	dir := t.TempDir()

	resetFlags()
	jsonlSeed = 42
	jsonlOutput = dir + "/a.jsonl"
	execute(t, "jsonl", "3", "id=uuid")
	a := readFile(t, dir+"/a.jsonl")

	resetFlags()
	jsonlSeed = 42
	jsonlOutput = dir + "/b.jsonl"
	execute(t, "jsonl", "3", "id=uuid")
	b := readFile(t, dir+"/b.jsonl")

	if a != b {
		t.Error("JSONL not reproducible")
	}
}

func TestTemplateEcommerce(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	tmplOutput = dir

	_, err := execute(t, "template", "ecommerce")
	if err != nil {
		t.Fatalf("template ecommerce: %v", err)
	}

	for _, f := range []string{"customers.sql", "products.sql", "orders.sql", "orders_csv.csv"} {
		content := readFile(t, dir+"/"+f)
		if len(content) == 0 {
			t.Errorf("empty file: %s", f)
		}
	}
}

func TestTemplateWeblogs(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	tmplOutput = dir

	_, err := execute(t, "template", "weblogs")
	if err != nil {
		t.Fatalf("template weblogs: %v", err)
	}

	for _, f := range []string{"access_logs.csv", "api_errors.jsonl"} {
		content := readFile(t, dir+"/"+f)
		if len(content) == 0 {
			t.Errorf("empty file: %s", f)
		}
	}
}

func TestTemplateIoT(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	tmplOutput = dir

	_, err := execute(t, "template", "iot")
	if err != nil {
		t.Fatalf("template iot: %v", err)
	}

	for _, f := range []string{"devices.sql", "sensor_readings.csv", "alerts.jsonl"} {
		content := readFile(t, dir+"/"+f)
		if len(content) == 0 {
			t.Errorf("empty file: %s", f)
		}
	}
}

func TestTemplateWithRowsOverride(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	tmplOutput = dir
	tmplOverrides = "customers:5,orders:10"

	_, err := execute(t, "template", "ecommerce")
	if err != nil {
		t.Fatalf("template with overrides: %v", err)
	}

	readFile(t, dir+"/customers.sql")
	readFile(t, dir+"/orders.sql")
}

func TestTemplateUnknown(t *testing.T) {
	resetFlags()
	_, err := execute(t, "template", "nonexistent")
	if err == nil {
		t.Error("expected error for unknown template")
	}
}

func TestTemplateMissingArg(t *testing.T) {
	resetFlags()
	_, err := execute(t, "template")
	if err == nil {
		t.Error("expected error for missing template name")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}
