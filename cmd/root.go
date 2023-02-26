package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "golling",
		Short: "golling update golang to the latest version.",
		Long: `golling updates golang to the latest version in /usr/local/go.

golling start update if golang is not up to date. By default, golling
checks /usr/local/go. If golang is not on the system, golling install the
latest golang in /usr/local/go. golling does not support Windows.`,
		Example: "  sudo golling update",
	}
}

// Execute run golling process.
func Execute() int {
	if isWindows() {
		fmt.Printf("%s does not support windows.", Name)
		return 1
	}

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newUpdateCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newCompletionCmd())

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}
