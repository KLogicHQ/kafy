package cmd

import (
	"fmt"
	"os"
	"strings"

	"kafy/internal/output"

	"github.com/spf13/cobra"
)

const version = "0.0.7"

var (
	outputFormat    string
	clusterOverride string
	rootCmd         = &cobra.Command{
		Use:     "kafy <command> <subcommand> [flags]",
		Version: version,
		Short:   "Kafka Productivity CLI - A Unified CLI for Kafka (v" + version + ")",
		Long: `kafy is a comprehensive CLI tool for managing Kafka clusters.
It provides a kubectl-inspired interface for working with topics, consumer groups,
producers, consumers, and cluster administration.

`,
		// Usage:
		//   kafy [command] [subcommand] [flags] [options]

		// CORE COMMANDS:
		//   config        Manage cluster configurations and contexts
		//   topics        Manage Kafka topics (create, list, describe, delete)
		//   groups        Manage consumer groups and offsets
		//   produce       Produce messages to topics
		//   consume       Consume messages from topics
		//   brokers       Inspect and manage Kafka brokers
		//   offsets       View and manage topic/partition offsets
		//   health        Check cluster health and connectivity
		//   tail          Tail messages in real-time (like tail -f)
		//   util          Utility commands for cluster administration

		// ADDITIONAL COMMANDS:
		//   completion    Generate the autocompletion script for the specified shell
		//   help          Help about any command
		//   version       Show version information

		// EXAMPLES:
		//   $ kafy config add my-cluster --bootstrap kafka.example.com:9092
		//   $ kafy topics list
		//   $ kafy topics create orders --partitions 3 --replication 2
		//   $ kafy produce orders --count 10
		//   $ kafy consume orders --from-beginning --limit 5
		//   $ kafy tail orders
		//   $ kafy groups list
		//   $ kafy brokers list`,
	}
)

func Execute() error {
	// Walk through all commands and wrap their RunE functions
	wrapAllCommands(rootCmd)
	return rootCmd.Execute()
}

// wrapAllCommands recursively wraps all RunE functions in the command tree
func wrapAllCommands(cmd *cobra.Command) {
	if cmd.RunE != nil {
		originalRunE := cmd.RunE
		cmd.RunE = wrapRunE(originalRunE)
	}
	for _, subCmd := range cmd.Commands() {
		wrapAllCommands(subCmd)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	rootCmd.PersistentFlags().StringVarP(&clusterOverride, "cluster", "c", "", "Use specified cluster instead of current context")

	// Add completion for output format
	rootCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json", "yaml"}, cobra.ShellCompDirectiveDefault
	})

	// Enable completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = false
	rootCmd.CompletionOptions.HiddenDefaultCmd = false

	// Add subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(topicsCmd)
	rootCmd.AddCommand(groupsCmd)
	rootCmd.AddCommand(produceCmd)
	rootCmd.AddCommand(consumeCmd)
	rootCmd.AddCommand(cpCmd)
	rootCmd.AddCommand(brokersCmd)
	rootCmd.AddCommand(offsetsCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(tailCmd)
	rootCmd.AddCommand(utilCmd)
}

func getFormatter() *output.Formatter {
	return output.NewFormatter(outputFormat)
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// isUsageError checks if an error is related to command usage (args/flags) vs operational errors
func isUsageError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "arg(s)") ||
		strings.Contains(errMsg, "unknown command") ||
		strings.Contains(errMsg, "unknown flag") ||
		strings.Contains(errMsg, "invalid argument") ||
		strings.Contains(errMsg, "requires at least") ||
		strings.Contains(errMsg, "accepts") ||
		strings.Contains(errMsg, "required flag") ||
		strings.Contains(errMsg, "flag needs")
}

// wrapRunE wraps a command's RunE to intelligently handle SilenceUsage
// For operational errors, it silences usage. For argument errors, usage is shown.
func wrapRunE(runE func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := runE(cmd, args)
		if err != nil && !isUsageError(err) {
			// For operational errors, silence usage
			cmd.SilenceUsage = true
		}
		return err
	}
}
