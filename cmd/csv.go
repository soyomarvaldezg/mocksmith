package cmd

import (
	"fmt"
	"strconv"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/soyomarvaldezg/mocksmith/internal/render"
	"github.com/soyomarvaldezg/mocksmith/internal/schema"
	"github.com/spf13/cobra"
)

var (
	csvSeed          int64
	csvDuplicateRate float64
	csvNullRate      float64
	csvCorruptRate   float64
	csvOutlierRate   float64
	csvOutput        string
)

var csvCmd = &cobra.Command{
	Use:   "csv <rows> <columns...>",
	Short: "Generate CSV data",
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

		faker := gofakeit.New(uint64(csvSeed))

		w, closeW, err := openOutput(csvOutput)
		if err != nil {
			return err
		}
		defer closeW()

		return render.CSV(
			faker,
			w,
			rows,
			columns,
			csvDuplicateRate,
			csvNullRate,
			csvCorruptRate,
			csvOutlierRate,
		)
	},
}

func init() {
	rootCmd.AddCommand(csvCmd)
	csvCmd.Flags().Int64Var(&csvSeed, "seed", 0, "seed for reproducible output (0 = random)")
	csvCmd.Flags().Float64Var(&csvDuplicateRate, "duplicate-rate", 0, "fraction of rows to duplicate (0.0-1.0)")
	csvCmd.Flags().Float64Var(&csvNullRate, "null", 0, "fraction of values to leave empty (0.0-1.0)")
	csvCmd.Flags().Float64Var(&csvCorruptRate, "corrupt", 0, "fraction of non-empty values to corrupt (0.0-1.0)")
	csvCmd.Flags().Float64Var(&csvOutlierRate, "outlier", 0, "fraction of non-empty values to replace with outliers (0.0-1.0)")
	csvCmd.Flags().StringVarP(&csvOutput, "output", "o", "", "write output to file instead of stdout")
}
