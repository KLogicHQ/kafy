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

var topicsPartitionsCmd = &cobra.Command{
        Use:   "partitions [topic]",
        Short: "Show partition details for all topics or a specific topic",
        Args:  cobra.MaximumNArgs(1),
        ValidArgsFunction: completeTopics,
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                metadata, err := client.DumpMetadata()
                if err != nil {
                        return err
                }

                // Filter for specific topic if provided
                var targetTopic string
                if len(args) > 0 {
                        targetTopic = args[0]
                }

                return displayPartitionInfo(metadata, targetTopic)
        },
}

func displayPartitionInfo(metadata interface{}, targetTopic string) error {
        data := metadata.(map[string]interface{})
        topics := data["topics"].([]map[string]interface{})
        
        if len(topics) == 0 {
                fmt.Println("No topics found")
                return nil
        }

        headers := []string{"TOPIC", "PARTITION", "LEADER", "REPLICAS", "ISRS"}
        var rows [][]string

        for _, topic := range topics {
                topicName := fmt.Sprintf("%v", topic["name"])
                
                // Skip if we're looking for a specific topic and this isn't it
                if targetTopic != "" && topicName != targetTopic {
                        continue
                }
                
                partitions := topic["partitions"].([]map[string]interface{})
                
                for _, partition := range partitions {
                        replicas := partition["replicas"].([]int32)
                        isrs := partition["isrs"].([]int32)
                        
                        replicaStrs := make([]string, len(replicas))
                        for i, r := range replicas {
                                replicaStrs[i] = fmt.Sprintf("%d", r)
                        }
                        
                        isrStrs := make([]string, len(isrs))
                        for i, isr := range isrs {
                                isrStrs[i] = fmt.Sprintf("%d", isr)
                        }
                        
                        row := []string{
                                topicName,
                                fmt.Sprintf("%v", partition["id"]),
                                fmt.Sprintf("%v", partition["leader"]),
                                "[" + strings.Join(replicaStrs, ",") + "]",
                                "[" + strings.Join(isrStrs, ",") + "]",
                        }
                        rows = append(rows, row)
                }
        }

        if len(rows) == 0 {
                if targetTopic != "" {
                        fmt.Printf("Topic '%s' not found\n", targetTopic)
                } else {
                        fmt.Println("No partition information found")
                }
                return nil
        }

        getFormatter().OutputTable(headers, rows)
        return nil
}

// Topic config commands
var topicsConfigsCmd = &cobra.Command{
        Use:   "configs",
        Short: "Manage topic configurations",
        Long:  "Commands for managing topic-level configurations",
}

var topicsConfigsListCmd = &cobra.Command{
        Use:   "list",
        Short: "List configurations for all topics",
        Args:  cobra.NoArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                // Get all topics first
                topics, err := client.ListTopics()
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                
                if outputFormat == "table" {
                        // For table output, show all topic configs in a structured format
                        headers := []string{"Topic", "Config Key", "Value"}
                        var rows [][]string
                        
                        for _, topic := range topics {
                                configs, err := client.GetTopicConfig(topic.Name)
                                if err != nil {
                                        continue // Skip topics we can't access
                                }
                                
                                // configs is map[string]string from the client method
                                for key, value := range configs {
                                        rows = append(rows, []string{topic.Name, key, value})
                                }
                        }
                        
                        formatter.OutputTable(headers, rows)
                        return nil
                } else {
                        // For JSON/YAML, collect all configs in structured format
                        allConfigs := make(map[string]interface{})
                        
                        for _, topic := range topics {
                                configs, err := client.GetTopicConfig(topic.Name)
                                if err != nil {
                                        continue // Skip topics we can't access
                                }
                                allConfigs[topic.Name] = configs
                        }
                        
                        return formatter.Output(allConfigs)
                }
        },
}

var topicsConfigsGetCmd = &cobra.Command{
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

var topicsConfigsSetCmd = &cobra.Command{
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

var topicsConfigsDeleteCmd = &cobra.Command{
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
        topicsCmd.AddCommand(topicsPartitionsCmd)
        topicsCmd.AddCommand(topicsConfigsCmd)

        // Add completion support
        topicsDescribeCmd.ValidArgsFunction = completeTopics
        topicsDeleteCmd.ValidArgsFunction = completeTopics
        topicsAlterCmd.ValidArgsFunction = completeTopics
        topicsPartitionsCmd.ValidArgsFunction = completeTopics
        topicsConfigsGetCmd.ValidArgsFunction = completeTopics
        topicsConfigsSetCmd.ValidArgsFunction = completeTopics
        topicsConfigsDeleteCmd.ValidArgsFunction = completeTopics

        // Add config subcommands
        topicsConfigsCmd.AddCommand(topicsConfigsListCmd)
        topicsConfigsCmd.AddCommand(topicsConfigsGetCmd)
        topicsConfigsCmd.AddCommand(topicsConfigsSetCmd)
        topicsConfigsCmd.AddCommand(topicsConfigsDeleteCmd)
        
        // Add completion support for config list
        topicsConfigsListCmd.ValidArgsFunction = completeTopics

        // Add flags
        topicsCreateCmd.Flags().Int("partitions", 1, "Number of partitions")
        topicsCreateCmd.Flags().Int("replication", 1, "Replication factor")
        topicsDeleteCmd.Flags().Bool("force", false, "Skip confirmation")
        topicsAlterCmd.Flags().Int("partitions", 0, "New number of partitions")
}