package cmd

import (
	"fmt"
	"strconv"

	"codeberg.org/soyomarvaldezg/mocksmith/internal/render"
	"codeberg.org/soyomarvaldezg/mocksmith/internal/schema"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/spf13/cobra"
)

var (
	jsonlSeed          int64
	jsonlDuplicateRate float64
	jsonlNullRate      float64
	jsonlCorruptRate   float64
	jsonlOutlierRate   float64
	jsonlOutput        string
)

var jsonlCmd = &cobra.Command{
	Use:   "jsonl <rows> <columns...>",
	Short: "Generate JSON Lines data",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		rows, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("rows must be an integer")
		}

		columns, err := schema.ParseColumns(args[1:])
		if err != nil {
			return err
		}

		faker := gofakeit.New(uint64(jsonlSeed))

		w, closeW, err := openOutput(jsonlOutput)
		if err != nil {
			return err
		}
		defer closeW()

		return render.JSONL(
			faker,
			w,
			rows,
			columns,
			jsonlDuplicateRate,
			jsonlNullRate,
			jsonlCorruptRate,
			jsonlOutlierRate,
		)
	},
}

func init() {
	rootCmd.AddCommand(jsonlCmd)
	jsonlCmd.Flags().Int64Var(&jsonlSeed, "seed", 0, "seed for reproducible output (0 = random)")
	jsonlCmd.Flags().Float64Var(&jsonlDuplicateRate, "duplicate-rate", 0, "fraction of rows to duplicate (0.0-1.0)")
	jsonlCmd.Flags().Float64Var(&jsonlNullRate, "null", 0, "fraction of values to set to null (0.0-1.0)")
	jsonlCmd.Flags().Float64Var(&jsonlCorruptRate, "corrupt", 0, "fraction of non-null values to corrupt (0.0-1.0)")
	jsonlCmd.Flags().Float64Var(&jsonlOutlierRate, "outlier", 0, "fraction of non-null values to replace with outliers (0.0-1.0)")
	jsonlCmd.Flags().StringVarP(&jsonlOutput, "output", "o", "", "write output to file instead of stdout")
}
