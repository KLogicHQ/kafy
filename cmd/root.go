package cmd

import (
        "fmt"
        "os"

        "kaf/internal/output"

        "github.com/spf13/cobra"
)

const version = "1.0.0"

var (
        outputFormat string
        rootCmd      = &cobra.Command{
                Use:     "kaf",
                Version: version,
                Short:   "Kafka Productivity CLI - A Unified CLI for Kafka",
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
