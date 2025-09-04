package cmd

import (
	"fmt"
	"os"

	"kaf/internal/output"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	rootCmd      = &cobra.Command{
		Use:   "kaf",
		Short: "Kafka Productivity CLI - A Unified CLI for Kafka",
		Long: `kaf is a comprehensive CLI tool for managing Kafka clusters.
It provides a kubectl-inspired interface for working with topics, consumer groups,
producers, consumers, and cluster administration.`,
		SilenceUsage: true,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")

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
