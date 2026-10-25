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
	sqlSeed          int64
	sqlDuplicateRate float64
	sqlNullRate      float64
	sqlCorruptRate   float64
	sqlOutlierRate   float64
	sqlDialect       string
	sqlOutput        string
)

var sqlCmd = &cobra.Command{
	Use:   "sql <table> <rows> <columns...>",
	Short: "Generate SQL INSERT statements",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		table := args[0]

		rows, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("rows must be an integer")
		}

		columns, err := schema.ParseColumns(args[2:])
		if err != nil {
			return err
		}

		dialect, err := render.ParseSQLDialect(sqlDialect)
		if err != nil {
			return err
		}

		faker := gofakeit.New(uint64(sqlSeed))

		w, closeW, err := openOutput(sqlOutput)
		if err != nil {
			return err
		}
		defer closeW()

		return render.SQLWithDialect(
			faker,
			w,
			table,
			rows,
			columns,
			sqlDuplicateRate,
			sqlNullRate,
			sqlCorruptRate,
			sqlOutlierRate,
			dialect,
		)
	},
}

func init() {
	rootCmd.AddCommand(sqlCmd)
	sqlCmd.Flags().Int64Var(&sqlSeed, "seed", 0, "seed for reproducible output (0 = random)")
	sqlCmd.Flags().Float64Var(&sqlDuplicateRate, "duplicate-rate", 0, "fraction of rows to duplicate (0.0-1.0)")
	sqlCmd.Flags().Float64Var(&sqlNullRate, "null", 0, "fraction of values to set to NULL (0.0-1.0)")
	sqlCmd.Flags().Float64Var(&sqlCorruptRate, "corrupt", 0, "fraction of non-null values to corrupt (0.0-1.0)")
	sqlCmd.Flags().Float64Var(&sqlOutlierRate, "outlier", 0, "fraction of non-null values to replace with outliers (0.0-1.0)")
	sqlCmd.Flags().StringVar(&sqlDialect, "dialect", "standard", "SQL dialect for string escaping (standard or mysql)")
	sqlCmd.Flags().StringVarP(&sqlOutput, "output", "o", "", "write output to file instead of stdout")
}
