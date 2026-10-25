package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mocksmith",
	Short: "Generate fake data for testing and demos",
	Long:  `Mocksmith generates fake datasets for SQL, CSV and JSONL outputs.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
