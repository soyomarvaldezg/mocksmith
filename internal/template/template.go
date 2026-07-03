package template

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"codeberg.org/soyomarvaldezg/mocksmith/internal/render"
	"codeberg.org/soyomarvaldezg/mocksmith/internal/schema"
	"github.com/brianvoe/gofakeit/v7"
	"gopkg.in/yaml.v3"
)

type Template struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Seed        int64       `yaml:"seed"`
	Tables      []TableDef  `yaml:"tables"`
}

type TableDef struct {
	Name    string      `yaml:"name"`
	Rows    int         `yaml:"rows"`
	Format  string      `yaml:"format"`
	Columns []ColumnDef `yaml:"columns"`
}

type ColumnDef struct {
	Name      string `yaml:"name"`
	Generator string `yaml:"generator"`
}

func Load(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseData(data)
}

func ParseData(data []byte) (*Template, error) {
	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("invalid template YAML: %w", err)
	}
	return &tmpl, nil
}

type GenerateOptions struct {
	OutputDir string
	Overrides map[string]int
}

func (tmpl *Template) Generate(opts GenerateOptions) error {
	seed := uint64(tmpl.Seed)
	faker := gofakeit.New(seed)

	for _, table := range tmpl.Tables {
		rows := table.Rows
		if n, ok := opts.Overrides[table.Name]; ok {
			rows = n
		}

		args := make([]string, 0, len(table.Columns))
		for _, c := range table.Columns {
			gen := c.Generator
			// Resolve ref: paths relative to output directory
			if opts.OutputDir != "" && strings.HasPrefix(gen, "ref:") {
				refPath := strings.TrimPrefix(gen, "ref:")
				gen = "ref:" + filepath.Join(opts.OutputDir, refPath)
			}
			args = append(args, c.Name+"="+gen)
		}

		cols, err := schema.ParseColumns(args)
		if err != nil {
			return fmt.Errorf("table %q: %w", table.Name, err)
		}

		ext := formatExt(table.Format)
		filename := table.Name + "." + ext
		if opts.OutputDir != "" {
			filename = filepath.Join(opts.OutputDir, filename)
			if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
				return fmt.Errorf("table %q: %w", table.Name, err)
			}
		}

		w, closeW, err := openFile(filename)
		if err != nil {
			return fmt.Errorf("table %q: %w", table.Name, err)
		}

		var genErr error
		switch table.Format {
		case "sql":
			genErr = render.SQL(faker, w, table.Name, rows, cols, 0, 0, 0, 0)
		case "csv":
			genErr = render.CSV(faker, w, rows, cols, 0, 0, 0, 0)
		case "jsonl":
			genErr = render.JSONL(faker, w, rows, cols, 0, 0, 0, 0)
		default:
			closeW()
			return fmt.Errorf("table %q: unsupported format %q", table.Name, table.Format)
		}

		if genErr != nil {
			closeW()
			return fmt.Errorf("table %q: %w", table.Name, genErr)
		}

		if err := closeW(); err != nil {
			return fmt.Errorf("table %q: %w", table.Name, err)
		}

		fmt.Fprintf(os.Stderr, "  ✓ %s (%d rows)\n", filename, rows)
	}

	return nil
}

func formatExt(format string) string {
	switch format {
	case "sql", "csv", "jsonl":
		return format
	default:
		return "out"
	}
}

func openFile(path string) (io.Writer, func() error, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return f, f.Close, nil
}

func ParseOverrides(s string) map[string]int {
	if s == "" {
		return nil
	}
	m := make(map[string]int)
	for part := range strings.SplitSeq(s, ",") {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			if n, err := strconv.Atoi(kv[1]); err == nil {
				m[strings.TrimSpace(kv[0])] = n
			}
		}
	}
	return m
}
