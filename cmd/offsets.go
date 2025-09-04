package cmd

import (
        "fmt"
        "strconv"

        "github.com/spf13/cobra"
        "kaf/config"
        "kaf/internal/kafka"
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
                
                cfg, err := config.LoadConfig()
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
        Short: "Reset topic offsets",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                toEarliest, _ := cmd.Flags().GetBool("to-earliest")
                toLatest, _ := cmd.Flags().GetBool("to-latest")
                toTimestamp, _ := cmd.Flags().GetString("to-timestamp")
                
                if !toEarliest && !toLatest && toTimestamp == "" {
                        return fmt.Errorf("must specify one of --to-earliest, --to-latest, or --to-timestamp")
                }
                
                // For now, return a message indicating this feature needs more implementation
                fmt.Printf("Offset reset for topic '%s' - feature requires additional implementation\n", topicName)
                fmt.Println("This would reset consumer offsets based on the specified strategy")
                return nil
        },
}

func init() {
        offsetsCmd.AddCommand(offsetsShowCmd)
        offsetsCmd.AddCommand(offsetsResetCmd)

        // Add flags for reset command
        offsetsResetCmd.Flags().Bool("to-earliest", false, "Reset to earliest offset")
        offsetsResetCmd.Flags().Bool("to-latest", false, "Reset to latest offset")
        offsetsResetCmd.Flags().String("to-timestamp", "", "Reset to timestamp")
}