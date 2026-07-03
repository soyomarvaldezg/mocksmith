package cmd

import (
	"bufio"
	"io"
	"os"
)

func openOutput(path string) (io.Writer, func() error, error) {
	if path == "" {
		w := bufio.NewWriter(os.Stdout)
		return w, w.Flush, nil
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}

	w := bufio.NewWriter(f)
	return w, func() error {
		if err := w.Flush(); err != nil {
			f.Close()
			return err
		}
		return f.Close()
	}, nil
}
