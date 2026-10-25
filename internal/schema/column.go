package schema

import (
	"fmt"
	"strings"
)

type Column struct {
	Name      string
	Generator string
	Path      []string // dotted path: "address.city" → ["address", "city"]
	ArrayKey  string   // non-empty if this column belongs to an array (e.g. "orders")
}

func ParseColumns(args []string) ([]Column, error) {
	columns := make([]Column, 0, len(args))

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)

		if len(parts) != 2 {
			return nil, fmt.Errorf(
				"invalid column definition: %s (expected name=generator)",
				arg,
			)
		}

		name := parts[0]
		gen := parts[1]

		if name == "" {
			return nil, fmt.Errorf("invalid column definition: %s (empty name)", arg)
		}

		col := Column{
			Name:      name,
			Generator: gen,
		}

		// Parse path from dotted name, handling array notation
		// "address.city" → path ["address","city"]
		// "orders[].id" → path ["orders","id"], ArrayKey="orders"
		segments := strings.Split(name, ".")
		path := make([]string, 0, len(segments))

		for _, seg := range segments {
			if key, ok := strings.CutSuffix(seg, "[]"); ok {
				if key == "" {
					return nil, fmt.Errorf("invalid column definition: %s (empty array name)", arg)
				}
				col.ArrayKey = key
				// don't add "[]" to path, just remember it's an array
			} else {
				if seg == "" {
					return nil, fmt.Errorf("invalid column definition: %s (empty path segment)", arg)
				}
				path = append(path, seg)
			}
		}

		if col.ArrayKey != "" && len(path) == 0 {
			return nil, fmt.Errorf("invalid column definition: %s (array fields need a name, e.g. orders[].id)", arg)
		}

		col.Path = path
		columns = append(columns, col)
	}

	return columns, nil
}
