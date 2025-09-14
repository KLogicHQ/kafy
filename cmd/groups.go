package cmd

import (
        "fmt"
        "strconv"
        "strings"

        "github.com/spf13/cobra"
        kafkaClient "kafy/internal/kafka"
)

var groupsCmd = &cobra.Command{
        Use:   "groups",
        Short: "Manage consumer groups",
        Long:  "Commands for managing Kafka consumer groups",
}

var groupsListCmd = &cobra.Command{
        Use:   "list",
        Short: "List all consumer groups",
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
                if err != nil {
                        return err
                }

                groups, err := client.ListConsumerGroups()
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                headers := []string{"Group ID", "State", "Members"}
                var rows [][]string

                for _, group := range groups {
                        rows = append(rows, []string{
                                group.GroupID,
                                group.State,
                                strconv.Itoa(group.MemberCount),
                        })
                }

                formatter.OutputTable(headers, rows)
                return nil
        },
}

var groupsDescribeCmd = &cobra.Command{
        Use:   "describe <group>",
        Short: "Show group details (members, offsets)",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                groupID := args[0]
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
                if err != nil {
                        return err
                }

                groupInfo, err := client.DescribeConsumerGroup(groupID)
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                if formatter.Format == "table" {
                        headers := []string{"Property", "Value"}
                        rows := [][]string{
                                {"Group ID", groupInfo.GroupID},
                                {"State", groupInfo.State},
                                {"Members", strconv.Itoa(len(groupInfo.Members))},
                        }
                        
                        if len(groupInfo.Members) > 0 {
                                rows = append(rows, []string{"Member Details", ""})
                                for i, member := range groupInfo.Members {
                                        rows = append(rows, []string{fmt.Sprintf("  Member %d ID", i+1), member.MemberID})
                                        rows = append(rows, []string{fmt.Sprintf("  Member %d Client", i+1), member.ClientID})
                                        rows = append(rows, []string{fmt.Sprintf("  Member %d Host", i+1), member.Host})
                                }
                        }
                        
                        formatter.OutputTable(headers, rows)
                        return nil
                }
                
                return formatter.Output(groupInfo)
        },
}

var groupsLagCmd = &cobra.Command{
        Use:   "lag <group>",
        Short: "Show lag metrics per partition",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                groupID := args[0]
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
                if err != nil {
                        return err
                }

                lag, err := client.GetConsumerGroupLag(groupID)
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                if len(lag) == 0 {
                        fmt.Printf("No lag data available for group '%s'\n", groupID)
                        return nil
                }
                
                if formatter.Format == "table" {
                        headers := []string{"Topic", "Partition", "Lag"}
                        var rows [][]string
                        totalLagPerTopic := make(map[string]int64)
                        var totalLag int64
                        
                        for topic, partitions := range lag {
                                for partition, lagValue := range partitions {
                                        rows = append(rows, []string{
                                                topic,
                                                strconv.Itoa(int(partition)),
                                                strconv.Itoa(int(lagValue)),
                                        })
                                        totalLagPerTopic[topic] += lagValue
                                        totalLag += lagValue
                                }
                        }
                        
                        formatter.OutputTable(headers, rows)
                        
                        // Display summary with total lag per topic and overall total
                        fmt.Println("\nLag Summary:")
                        summaryHeaders := []string{"Topic", "Total Lag"}
                        var summaryRows [][]string
                        
                        for topic, topicLag := range totalLagPerTopic {
                                summaryRows = append(summaryRows, []string{
                                        topic,
                                        strconv.Itoa(int(topicLag)),
                                })
                        }
                        
                        // Add overall total
                        summaryRows = append(summaryRows, []string{
                                "TOTAL",
                                strconv.Itoa(int(totalLag)),
                        })
                        
                        formatter.OutputTable(summaryHeaders, summaryRows)
                        return nil
                }
                
                return formatter.Output(lag)
        },
}

var groupsResetCmd = &cobra.Command{
        Use:   "reset <group>",
        Short: "Reset consumer group offsets",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                groupID := args[0]
                toEarliest, _ := cmd.Flags().GetBool("to-earliest")
                toLatest, _ := cmd.Flags().GetBool("to-latest")
                toTimestamp, _ := cmd.Flags().GetString("to-timestamp")
                
                if !toEarliest && !toLatest && toTimestamp == "" {
                        return fmt.Errorf("must specify one of --to-earliest, --to-latest, or --to-timestamp")
                }
                
                // For now, return a message indicating this feature needs more implementation
                fmt.Printf("Consumer group offset reset for '%s' - feature requires additional implementation\n", groupID)
                fmt.Println("This would reset offsets based on the specified strategy")
                return nil
        },
}

var groupsDeleteCmd = &cobra.Command{
        Use:   "delete <group>",
        Short: "Delete a consumer group",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                groupID := args[0]
                force, _ := cmd.Flags().GetBool("force")
                
                if !force {
                        fmt.Printf("Are you sure you want to delete consumer group '%s'? (y/N): ", groupID)
                        var response string
                        fmt.Scanln(&response)
                        if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
                                fmt.Println("Cancelled")
                                return nil
                        }
                }
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
                if err != nil {
                        return err
                }

                if err := client.DeleteConsumerGroup(groupID); err != nil {
                        return err
                }

                fmt.Printf("Consumer group '%s' deleted successfully\n", groupID)
                return nil
        },
}

func init() {
        groupsCmd.AddCommand(groupsListCmd)
        groupsCmd.AddCommand(groupsDescribeCmd)
        groupsCmd.AddCommand(groupsLagCmd)
        groupsCmd.AddCommand(groupsResetCmd)
        groupsCmd.AddCommand(groupsDeleteCmd)

        // Add completion support
        groupsDescribeCmd.ValidArgsFunction = completeGroups
        groupsLagCmd.ValidArgsFunction = completeGroups
        groupsResetCmd.ValidArgsFunction = completeGroups
        groupsDeleteCmd.ValidArgsFunction = completeGroups

        // Add flags for reset command
        groupsResetCmd.Flags().Bool("to-earliest", false, "Reset to earliest offset")
        groupsResetCmd.Flags().Bool("to-latest", false, "Reset to latest offset")
        groupsResetCmd.Flags().String("to-timestamp", "", "Reset to timestamp")
        
        // Add flags for delete command
        groupsDeleteCmd.Flags().Bool("force", false, "Skip confirmation")
}