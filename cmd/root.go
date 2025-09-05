package cmd

import (
        "fmt"
        "os"

        "kaf/internal/output"

        "github.com/spf13/cobra"
)

const version = "0.0.2"

var (
        outputFormat string
        rootCmd      = &cobra.Command{
                Use:     "kaf <command> <subcommand> [flags]",
                Version: version,
                Short:   "Kafka Productivity CLI - A Unified CLI for Kafka",
                Long: `kaf is a comprehensive CLI tool for managing Kafka clusters.
It provides a kubectl-inspired interface for working with topics, consumer groups,
producers, consumers, and cluster administration.

`,
// Usage:
//   kaf [command] [subcommand] [flags] [options]

// CORE COMMANDS:
//   config        Manage cluster configurations and contexts
//   topics        Manage Kafka topics (create, list, describe, delete)
//   groups        Manage consumer groups and offsets
//   produce       Produce messages to topics
//   consume       Consume messages from topics
//   brokers       Inspect and manage Kafka brokers
//   offsets       View and manage topic/partition offsets
//   health        Check cluster health and connectivity
//   util          Utility commands for cluster administration

// ADDITIONAL COMMANDS:
//   completion    Generate the autocompletion script for the specified shell
//   help          Help about any command
//   version       Show version information

// EXAMPLES:
//   $ kaf config add my-cluster --bootstrap kafka.example.com:9092
//   $ kaf topics list
//   $ kaf topics create orders --partitions 3 --replication 2
//   $ kaf produce orders --count 10
//   $ kaf consume orders --from-beginning --limit 5
//   $ kaf groups list
//   $ kaf brokers list`,
                SilenceUsage: true,
        }
)

func Execute() error {
        return rootCmd.Execute()
}

func init() {
        rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")

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
        rootCmd.AddCommand(brokersCmd)
        rootCmd.AddCommand(offsetsCmd)
        rootCmd.AddCommand(healthCmd)
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
