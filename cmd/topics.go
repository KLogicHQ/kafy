package cmd

import (
        "context"
        "fmt"
        "sort"
        "strconv"
        "strings"
        "time"

        "github.com/confluentinc/confluent-kafka-go/v2/kafka"
        "github.com/spf13/cobra"
        kafkaClient "kkl/internal/kafka"
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
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
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
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
                if err != nil {
                        return err
                }

                topic, err := client.DescribeTopic(topicName)
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                if formatter.Format == "table" {
                        // Basic topic information
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
                        
                        // Display detailed partition information
                        if len(topic.PartitionDetails) > 0 {
                                fmt.Println("\nPartition Details:")
                                partHeaders := []string{"Partition", "Leader", "Replicas", "In-Sync Replicas", "IN SYNC"}
                                var partRows [][]string
                                
                                for _, partition := range topic.PartitionDetails {
                                        replicas := make([]string, len(partition.Replicas))
                                        for i, r := range partition.Replicas {
                                                replicas[i] = strconv.Itoa(int(r))
                                        }
                                        
                                        isr := make([]string, len(partition.Isr))
                                        for i, r := range partition.Isr {
                                                isr[i] = strconv.Itoa(int(r))
                                        }
                                        
                                        // Check if replicas and ISR arrays are the same
                                        inSync := "No"
                                        if len(partition.Replicas) == len(partition.Isr) {
                                                // Create maps to compare array contents regardless of order
                                                replicaSet := make(map[int32]bool)
                                                for _, r := range partition.Replicas {
                                                        replicaSet[r] = true
                                                }
                                                
                                                isrSet := make(map[int32]bool)
                                                for _, r := range partition.Isr {
                                                        isrSet[r] = true
                                                }
                                                
                                                // Check if both sets contain the same elements
                                                same := true
                                                for r := range replicaSet {
                                                        if !isrSet[r] {
                                                                same = false
                                                                break
                                                        }
                                                }
                                                if same {
                                                        inSync = "Yes"
                                                }
                                        }
                                        
                                        partRows = append(partRows, []string{
                                                strconv.Itoa(int(partition.ID)),
                                                strconv.Itoa(int(partition.Leader)),
                                                strings.Join(replicas, ","),
                                                strings.Join(isr, ","),
                                                inSync,
                                        })
                                }
                                
                                formatter.OutputTable(partHeaders, partRows)
                        }
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

                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
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

                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
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
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
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
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
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

        headers := []string{"TOPIC", "PARTITION", "LEADER", "REPLICAS", "ISRS", "INSYNC"}
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
                        
                        // Calculate insync status by comparing replicas and ISRs
                        insync := "no"
                        if len(replicas) == len(isrs) {
                                // Create maps for comparison
                                replicaSet := make(map[int32]bool)
                                for _, r := range replicas {
                                        replicaSet[r] = true
                                }
                                
                                allInSync := true
                                for _, isr := range isrs {
                                        if !replicaSet[isr] {
                                                allInSync = false
                                                break
                                        }
                                }
                                
                                if allInSync {
                                        insync = "yes"
                                }
                        }
                        
                        row := []string{
                                topicName,
                                fmt.Sprintf("%v", partition["id"]),
                                fmt.Sprintf("%v", partition["leader"]),
                                "[" + strings.Join(replicaStrs, ",") + "]",
                                "[" + strings.Join(isrStrs, ",") + "]",
                                insync,
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
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
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
                                
                                // Get all keys and sort them alphabetically
                                keys := make([]string, 0, len(configs))
                                for key := range configs {
                                        keys = append(keys, key)
                                }
                                sort.Strings(keys)
                                
                                // Add rows in alphabetical order for each topic
                                for _, key := range keys {
                                        rows = append(rows, []string{topic.Name, key, configs[key]})
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
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
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

                // Get all keys and sort them alphabetically
                keys := make([]string, 0, len(configs))
                for key := range configs {
                        keys = append(keys, key)
                }
                sort.Strings(keys)

                // Add rows in alphabetical order
                for _, key := range keys {
                        rows = append(rows, []string{key, configs[key], "topic"})
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
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
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
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
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

var topicsMovePartitionCmd = &cobra.Command{
        Use:   "move-partition <topic>",
        Short: "Move data from one partition to another partition within the same topic",
        Long: `Move all messages from a source partition to a destination partition.
This command reads all messages from the source partition and writes them to the destination partition,
preserving keys, values, and headers. Optionally, the source partition data can be deleted after successful copy.

Examples:
  kkl topics move-partition orders --source-partition 0 --dest-partition 3
  kkl topics move-partition users --source-partition 2 --dest-partition 1 --delete-source`,
        Args:              cobra.ExactArgs(1),
        ValidArgsFunction: completeTopics,
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                sourcePartition, _ := cmd.Flags().GetInt32("source-partition")
                destPartition, _ := cmd.Flags().GetInt32("dest-partition")
                deleteSource, _ := cmd.Flags().GetBool("delete-source")
                
                if sourcePartition == destPartition {
                        return fmt.Errorf("source and destination partitions cannot be the same")
                }
                
                // Confirm the operation if delete-source is enabled
                if deleteSource {
                        fmt.Printf("WARNING: This will move data from partition %d to partition %d and DELETE the source data.\n", sourcePartition, destPartition)
                        fmt.Printf("Are you sure you want to continue? (y/N): ")
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

                // Create consumer for reading from source partition
                groupID := fmt.Sprintf("kkl-move-partition-%d", time.Now().Unix())
                consumer, err := client.CreateConsumerWithOffset(groupID, "earliest")
                if err != nil {
                        return fmt.Errorf("failed to create consumer: %w", err)
                }
                defer consumer.Close()

                // Create producer for writing to destination partition
                producer, err := client.CreateProducer()
                if err != nil {
                        return fmt.Errorf("failed to create producer: %w", err)
                }
                defer producer.Close()

                // Get high watermark for the source partition to determine end point
                adminClient, err := client.CreateAdminClient()
                if err != nil {
                        return fmt.Errorf("failed to create admin client: %w", err)
                }
                defer adminClient.Close()
                
                ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
                defer cancel()
                
                watermarks, err := adminClient.QueryWatermarkOffsets(ctx, 
                        kafka.TopicPartition{
                                Topic:     &topicName,
                                Partition: sourcePartition,
                        }, kafka.SetAdminRequestTimeout(30*time.Second))
                if err != nil {
                        return fmt.Errorf("failed to query watermark offsets: %w", err)
                }
                
                highWatermark := watermarks.High
                lowWatermark := watermarks.Low
                totalMessages := highWatermark - lowWatermark
                
                if totalMessages == 0 {
                        fmt.Printf("Source partition %d is empty, nothing to move.\n", sourcePartition)
                        return nil
                }
                
                fmt.Printf("Found %d messages in source partition %d (offsets %d to %d)\n", totalMessages, sourcePartition, lowWatermark, highWatermark-1)
                
                // Manually assign the specific source partition with specific offset range
                topicPartitions := []kafka.TopicPartition{
                        {
                                Topic:     &topicName,
                                Partition: sourcePartition,
                                Offset:    lowWatermark, // Start from low watermark
                        },
                }
                
                err = consumer.Assign(topicPartitions)
                if err != nil {
                        return fmt.Errorf("failed to assign source partition: %w", err)
                }

                fmt.Printf("Moving data from partition %d to partition %d in topic '%s'...\n", sourcePartition, destPartition, topicName)
                
                messageCount := int64(0)
                for {
                        msg, err := consumer.ReadMessage(1000 * time.Millisecond)
                        if err != nil {
                                if kafkaErr, ok := err.(kafka.Error); ok {
                                        if kafkaErr.Code() == kafka.ErrTimedOut {
                                                // No more messages available
                                                break
                                        }
                                }
                                return fmt.Errorf("failed to read message: %w", err)
                        }

                        // Check if we've reached the high watermark snapshot
                        if msg.TopicPartition.Offset >= highWatermark {
                                fmt.Printf("Reached high watermark (offset %d), completing move.\n", highWatermark)
                                break
                        }

                        // Produce message to destination partition, preserving timestamp and headers
                        destMessage := &kafka.Message{
                                TopicPartition: kafka.TopicPartition{
                                        Topic:     &topicName,
                                        Partition: destPartition,
                                },
                                Key:       msg.Key,
                                Value:     msg.Value,
                                Headers:   msg.Headers,
                                Timestamp: msg.Timestamp, // Preserve original timestamp
                        }

                        // Create delivery channel for this specific message
                        deliveryChan := make(chan kafka.Event)
                        err = producer.Produce(destMessage, deliveryChan)
                        if err != nil {
                                return fmt.Errorf("failed to produce message to destination partition: %w", err)
                        }

                        // Wait for delivery confirmation
                        e := <-deliveryChan
                        m := e.(*kafka.Message)
                        if m.TopicPartition.Error != nil {
                                return fmt.Errorf("delivery failed for message at offset %d: %w", msg.TopicPartition.Offset, m.TopicPartition.Error)
                        }

                        messageCount++
                        if messageCount%100 == 0 {
                                fmt.Printf("Moved %d/%d messages...\n", messageCount, totalMessages)
                        }
                }

                // Wait for all messages to be delivered
                producer.Flush(30000)
                
                if deleteSource {
                        return fmt.Errorf("--delete-source is not supported: Kafka does not provide safe APIs for deleting data from specific partitions. Use topic retention policies or manual cleanup tools instead")
                }

                fmt.Printf("\nPartition move completed successfully!\n")
                fmt.Printf("Total messages moved: %d/%d\n", messageCount, totalMessages)
                fmt.Printf("Data copied from partition %d to partition %d in topic '%s'\n", sourcePartition, destPartition, topicName)
                fmt.Printf("NOTE: Original data in source partition %d remains untouched for safety.\n", sourcePartition)
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
        topicsCmd.AddCommand(topicsMovePartitionCmd)

        // Add completion support
        topicsDescribeCmd.ValidArgsFunction = completeTopics
        topicsDeleteCmd.ValidArgsFunction = completeTopics
        topicsAlterCmd.ValidArgsFunction = completeTopics
        topicsPartitionsCmd.ValidArgsFunction = completeTopics
        topicsConfigsGetCmd.ValidArgsFunction = completeTopics
        topicsConfigsSetCmd.ValidArgsFunction = completeTopics
        topicsConfigsDeleteCmd.ValidArgsFunction = completeTopics
        topicsMovePartitionCmd.ValidArgsFunction = completeTopics

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
        
        // Move partition flags
        topicsMovePartitionCmd.Flags().Int32("source-partition", -1, "Source partition number to move data from")
        topicsMovePartitionCmd.Flags().Int32("dest-partition", -1, "Destination partition number to move data to")
        topicsMovePartitionCmd.Flags().Bool("delete-source", false, "UNSUPPORTED: Attempt to delete source partition data (will error)")
        topicsMovePartitionCmd.MarkFlagRequired("source-partition")
        topicsMovePartitionCmd.MarkFlagRequired("dest-partition")
}