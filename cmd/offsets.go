package cmd

import (
        "fmt"
        "strconv"

        "github.com/spf13/cobra"
        "kkl/internal/kafka"
)

var offsetsCmd = &cobra.Command{
        Use:   "offsets",
        Short: "Manage topic offsets",
        Long:  "Commands for managing and inspecting topic offsets",
}

var offsetsShowCmd = &cobra.Command{
        Use:   "show <topic>",
        Short: "Show offsets per partition",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                offsets, err := client.GetTopicOffsets(topicName)
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                if formatter.Format == "table" {
                        headers := []string{"Partition", "Offset"}
                        var rows [][]string
                        for partition, offset := range offsets {
                                rows = append(rows, []string{
                                        strconv.Itoa(int(partition)),
                                        strconv.Itoa(int(offset)),
                                })
                        }
                        formatter.OutputTable(headers, rows)
                        return nil
                }
                
                return formatter.Output(offsets)
        },
}

var offsetsResetCmd = &cobra.Command{
        Use:   "reset <topic>",
        Short: "Reset all partitions to earliest",  
        Long:  "Reset offsets per partition. Use --to-earliest, --to-latest, or --to-timestamp",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                toEarliest, _ := cmd.Flags().GetBool("to-earliest")
                toLatest, _ := cmd.Flags().GetBool("to-latest")
                toTimestamp, _ := cmd.Flags().GetString("to-timestamp")
                
                if !toEarliest && !toLatest && toTimestamp == "" {
                        return fmt.Errorf("must specify one of --to-earliest, --to-latest, or --to-timestamp")
                }
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                var resetType string
                if toEarliest {
                        resetType = "earliest"
                } else if toLatest {
                        resetType = "latest"
                } else {
                        resetType = toTimestamp
                }

                if err := client.ResetTopicOffsets(topicName, resetType); err != nil {
                        return err
                }

                fmt.Printf("Topic '%s' offsets reset to %s\n", topicName, resetType)
                return nil
        },
}

func init() {
        offsetsCmd.AddCommand(offsetsShowCmd)
        offsetsCmd.AddCommand(offsetsResetCmd)

        // Add completion support
        offsetsShowCmd.ValidArgsFunction = completeTopics
        offsetsResetCmd.ValidArgsFunction = completeTopics

        // Add flags for reset command
        offsetsResetCmd.Flags().Bool("to-earliest", false, "Reset to earliest offset")
        offsetsResetCmd.Flags().Bool("to-latest", false, "Reset to latest offset")
        offsetsResetCmd.Flags().String("to-timestamp", "", "Reset to timestamp")
}