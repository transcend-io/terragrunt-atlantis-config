package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	VERSION string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "terragrunt-atlantis-config",
	Short: "Generates Atlantis Config for Terragrunt projects",
	Long:  "Generates Atlantis Config for Terragrunt projects",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	VERSION = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
