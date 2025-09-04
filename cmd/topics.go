package cmd

import (
        "fmt"
        "strconv"
        "strings"

        "github.com/spf13/cobra"
        "kaf/config"
        "kaf/internal/kafka"
)

var topicsCmd = &cobra.Command{
        Use:   "topics",
        Short: "Manage Kafka topics",
        Long:  "Commands for managing Kafka topics",
}

var topicsListCmd = &cobra.Command{
        Use:   "list",
        Short: "List all topics",
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                topics, err := client.ListTopics()
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                headers := []string{"Name", "Partitions", "Replicas"}
                var rows [][]string

                for _, topic := range topics {
                        rows = append(rows, []string{
                                topic.Name,
                                strconv.Itoa(topic.Partitions),
                                strconv.Itoa(topic.Replicas),
                        })
                }

                formatter.OutputTable(headers, rows)
                return nil
        },
}

var topicsDescribeCmd = &cobra.Command{
        Use:   "describe <topic>",
        Short: "Show details of a topic",
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

                topic, err := client.DescribeTopic(topicName)
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                if formatter.Format == "table" {
                        headers := []string{"Property", "Value"}
                        rows := [][]string{
                                {"Name", topic.Name},
                                {"Partitions", strconv.Itoa(topic.Partitions)},
                                {"Replicas", strconv.Itoa(topic.Replicas)},
                        }
                        
                        // Add config entries if any
                        if len(topic.Config) > 0 {
                                rows = append(rows, []string{"Configs", ""})
                                for key, value := range topic.Config {
                                        rows = append(rows, []string{"  " + key, value})
                                }
                        }
                        
                        formatter.OutputTable(headers, rows)
                        return nil
                }
                
                return formatter.Output(topic)
        },
}

var topicsCreateCmd = &cobra.Command{
        Use:   "create <topic>",
        Short: "Create a topic",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                partitions, _ := cmd.Flags().GetInt("partitions")
                replication, _ := cmd.Flags().GetInt("replication")
                
                if partitions <= 0 {
                        return fmt.Errorf("partitions must be greater than 0")
                }
                if replication <= 0 {
                        return fmt.Errorf("replication factor must be greater than 0")
                }

                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                if err := client.CreateTopic(topicName, partitions, replication); err != nil {
                        return err
                }

                fmt.Printf("Topic '%s' created successfully\n", topicName)
                return nil
        },
}

var topicsDeleteCmd = &cobra.Command{
        Use:   "delete <topic>",
        Short: "Delete a topic (with safety confirmation)",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                force, _ := cmd.Flags().GetBool("force")
                
                if !force {
                        fmt.Printf("Are you sure you want to delete topic '%s'? (y/N): ", topicName)
                        var response string
                        fmt.Scanln(&response)
                        if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
                                fmt.Println("Cancelled")
                                return nil
                        }
                }

                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                if err := client.DeleteTopic(topicName); err != nil {
                        return err
                }

                fmt.Printf("Topic '%s' deleted successfully\n", topicName)
                return nil
        },
}

var topicsAlterCmd = &cobra.Command{
        Use:   "alter <topic>",
        Short: "Alter partitions or configs",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                partitions, _ := cmd.Flags().GetInt("partitions")
                
                if partitions <= 0 {
                        return fmt.Errorf("must specify --partitions with a value greater than 0")
                }
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                if err := client.AlterTopicPartitions(topicName, partitions); err != nil {
                        return err
                }

                fmt.Printf("Topic '%s' partitions updated to %d\n", topicName, partitions)
                return nil
        },
}

// Topic config commands
var topicsConfigCmd = &cobra.Command{
        Use:   "config",
        Short: "Manage topic configurations",
        Long:  "Commands for managing topic-level configurations",
}

var topicsConfigGetCmd = &cobra.Command{
        Use:   "get <topic>",
        Short: "Show topic-level configs",
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

                configs, err := client.GetTopicConfig(topicName)
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                headers := []string{"Config Key", "Value", "Source"}
                var rows [][]string

                for key, value := range configs {
                        rows = append(rows, []string{key, value, "topic"})
                }

                formatter.OutputTable(headers, rows)
                return nil
        },
}

var topicsConfigSetCmd = &cobra.Command{
        Use:   "set <topic> <key>=<value>",
        Short: "Set topic-level config",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                configPair := args[1]
                
                parts := strings.SplitN(configPair, "=", 2)
                if len(parts) != 2 {
                        return fmt.Errorf("config must be in format key=value")
                }
                
                key := parts[0]
                value := parts[1]
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                if err := client.SetTopicConfig(topicName, key, value); err != nil {
                        return err
                }

                fmt.Printf("Topic '%s' config '%s' set to '%s'\n", topicName, key, value)
                return nil
        },
}

var topicsConfigDeleteCmd = &cobra.Command{
        Use:   "delete <topic> <key>",
        Short: "Remove a config override",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                key := args[1]
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                if err := client.DeleteTopicConfig(topicName, key); err != nil {
                        return err
                }

                fmt.Printf("Topic '%s' config '%s' deleted\n", topicName, key)
                return nil
        },
}

func init() {
        topicsCmd.AddCommand(topicsListCmd)
        topicsCmd.AddCommand(topicsDescribeCmd)
        topicsCmd.AddCommand(topicsCreateCmd)
        topicsCmd.AddCommand(topicsDeleteCmd)
        topicsCmd.AddCommand(topicsAlterCmd)
        topicsCmd.AddCommand(topicsConfigCmd)

        // Add config subcommands
        topicsConfigCmd.AddCommand(topicsConfigGetCmd)
        topicsConfigCmd.AddCommand(topicsConfigSetCmd)
        topicsConfigCmd.AddCommand(topicsConfigDeleteCmd)

        // Add flags
        topicsCreateCmd.Flags().Int("partitions", 1, "Number of partitions")
        topicsCreateCmd.Flags().Int("replication", 1, "Replication factor")
        topicsDeleteCmd.Flags().Bool("force", false, "Skip confirmation")
        topicsAlterCmd.Flags().Int("partitions", 0, "New number of partitions")
}