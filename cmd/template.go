package cmd

import (
	_ "embed"
	"fmt"

	"github.com/soyomarvaldezg/mocksmith/internal/template"
	"github.com/spf13/cobra"
)

//go:embed templates/ecommerce.yaml
var ecommerceTemplate []byte

//go:embed templates/weblogs.yaml
var weblogsTemplate []byte

//go:embed templates/iot.yaml
var iotTemplate []byte

var (
	tmplOutput    string
	tmplOverrides string
)

var templateCmd = &cobra.Command{
	Use:   "template <name>",
	Short: "Generate a dataset from a predefined template",
	Long: `Generate a multi-table dataset from a predefined template.

Built-in templates:
  ecommerce  Customers, products, orders with foreign keys
  weblogs    Web server access logs and API errors
  iot        Devices, sensor readings, alerts

Usage:
  mocksmith template ecommerce
  mocksmith template ecommerce --output ./data/
  mocksmith template ecommerce --rows customers:1000,orders:5000`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		var data []byte
		switch name {
		case "ecommerce":
			data = ecommerceTemplate
		case "weblogs":
			data = weblogsTemplate
		case "iot":
			data = iotTemplate
		default:
			return fmt.Errorf("unknown template %q. Available: ecommerce, weblogs, iot", name)
		}

		tmpl, err := template.ParseData(data)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "▶ Generating %s: %s\n\n", tmpl.Name, tmpl.Description)

		overrides := template.ParseOverrides(tmplOverrides)

		return tmpl.Generate(template.GenerateOptions{
			OutputDir: tmplOutput,
			Overrides: overrides,
		})
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.Flags().StringVarP(&tmplOutput, "output", "o", "", "output directory for generated files")
	templateCmd.Flags().StringVar(&tmplOverrides, "rows", "", "override row counts (format: table:N,table:N)")
}
