package cmd

import (
        "fmt"
        "os"
        "os/signal"
        "syscall"
        "time"

        "github.com/confluentinc/confluent-kafka-go/v2/kafka"
        "github.com/spf13/cobra"
        kafkaClient "kkl/internal/kafka"
)

var cpCmd = &cobra.Command{
        Use:   "cp <source-topic> <destination-topic>",
        Short: "Copy messages from one topic to another",
        Long: `Copy messages from a source topic to a destination topic.
This command will consume messages from the source topic and produce them to the destination topic,
preserving keys, values, and headers.

Examples:
  kkl cp orders orders-backup
  kkl cp --from-beginning user-events user-events-copy
  kkl cp --limit 1000 transactions transactions-test
  kkl cp orders backup --begin-offset 100 --end-offset 500
  kkl cp events archive --begin-offset 1000`,
        Args:              cobra.ExactArgs(2),
        ValidArgsFunction: completeTopics,
        RunE: func(cmd *cobra.Command, args []string) error {
                sourceTopic := args[0]
                destTopic := args[1]
                fromBeginning, _ := cmd.Flags().GetBool("from-beginning")
                limit, _ := cmd.Flags().GetInt("limit")
                beginOffset, _ := cmd.Flags().GetInt64("begin-offset")
                endOffset, _ := cmd.Flags().GetInt64("end-offset")
                
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
                if err != nil {
                        return err
                }

                // Generate unique group ID for this copy operation
                group := fmt.Sprintf("kkl-cp-%d", time.Now().Unix())

                // Create consumer
                var offsetReset string
                if fromBeginning {
                        offsetReset = "earliest"
                } else {
                        offsetReset = "latest"
                }

                consumer, err := client.CreateConsumerWithOffset(group, offsetReset)
                if err != nil {
                        return err
                }
                defer consumer.Close()

                // Create producer
                producer, err := client.CreateProducer()
                if err != nil {
                        return err
                }
                defer producer.Close()

                // Check if offset-based copying is requested
                useOffsetBasedCopy := cmd.Flags().Changed("begin-offset")
                
                if useOffsetBasedCopy {
                        // Get topic metadata to determine partitions
                        adminClient, err := client.CreateAdminClient()
                        if err != nil {
                                return fmt.Errorf("failed to create admin client: %w", err)
                        }
                        defer adminClient.Close()

                        metadata, err := adminClient.GetMetadata(&sourceTopic, false, 5*1000)
                        if err != nil {
                                return fmt.Errorf("failed to get topic metadata: %w", err)
                        }

                        topic, exists := metadata.Topics[sourceTopic]
                        if !exists {
                                return fmt.Errorf("topic '%s' not found", sourceTopic)
                        }

                        // Assign specific partitions with begin offsets
                        var topicPartitions []kafka.TopicPartition
                        for _, partition := range topic.Partitions {
                                topicPartitions = append(topicPartitions, kafka.TopicPartition{
                                        Topic:     &sourceTopic,
                                        Partition: partition.ID,
                                        Offset:    kafka.Offset(beginOffset),
                                })
                        }

                        err = consumer.Assign(topicPartitions)
                        if err != nil {
                                return fmt.Errorf("failed to assign partitions: %w", err)
                        }

                        fmt.Printf("Starting to copy messages from '%s' to '%s' (offset-based)...\n", sourceTopic, destTopic)
                        fmt.Printf("Begin offset: %d", beginOffset)
                        if cmd.Flags().Changed("end-offset") {
                                fmt.Printf(", End offset: %d", endOffset)
                        }
                        fmt.Printf("\n")
                } else {
                        // Subscribe to source topic (existing behavior)
                        err = consumer.SubscribeTopics([]string{sourceTopic}, nil)
                        if err != nil {
                                return fmt.Errorf("failed to subscribe to source topic: %w", err)
                        }

                        fmt.Printf("Starting to copy messages from '%s' to '%s'...\n", sourceTopic, destTopic)
                }
                
                if limit > 0 {
                        fmt.Printf("Limiting to %d messages\n", limit)
                }

                // Setup signal handling for graceful shutdown
                sigChan := make(chan os.Signal, 1)
                signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

                messageCount := 0
                
        copyLoop:
                for {
                        select {
                        case <-sigChan:
                                fmt.Printf("\nReceived interrupt signal, stopping copy operation...\n")
                                break copyLoop
                        default:
                                msg, err := consumer.ReadMessage(100 * time.Millisecond)
                                if err != nil {
                                        // Check if it's just a timeout
                                        if kafkaErr, ok := err.(kafka.Error); ok {
                                                if kafkaErr.Code() == kafka.ErrTimedOut {
                                                        continue
                                                }
                                        }
                                        return fmt.Errorf("failed to read message: %w", err)
                                }

                                // Check if we've reached the end offset (for offset-based copying)
                                if useOffsetBasedCopy && cmd.Flags().Changed("end-offset") {
                                        if msg.TopicPartition.Offset >= kafka.Offset(endOffset) {
                                                fmt.Printf("Reached end offset %d, stopping copy operation...\n", endOffset)
                                                break copyLoop
                                        }
                                }

                                // Create message for destination topic
                                destMessage := &kafka.Message{
                                        TopicPartition: kafka.TopicPartition{
                                                Topic:     &destTopic,
                                                Partition: kafka.PartitionAny,
                                        },
                                        Key:     msg.Key,
                                        Value:   msg.Value,
                                        Headers: msg.Headers,
                                }

                                // Produce to destination topic
                                err = producer.Produce(destMessage, nil)
                                if err != nil {
                                        return fmt.Errorf("failed to produce message to destination: %w", err)
                                }

                                messageCount++
                                if messageCount%100 == 0 {
                                        fmt.Printf("Copied %d messages...\n", messageCount)
                                }

                                // Check if we've reached the limit
                                if limit > 0 && messageCount >= limit {
                                        break copyLoop
                                }
                        }
                }

                // Wait for all messages to be delivered
                producer.Flush(5000)

                fmt.Printf("\nCopy operation completed. Total messages copied: %d\n", messageCount)
                return nil
        },
}

func init() {
        cpCmd.Flags().Bool("from-beginning", false, "Start copying from the beginning of the topic")
        cpCmd.Flags().Int("limit", 0, "Maximum number of messages to copy (0 = unlimited)")
        cpCmd.Flags().Int64("begin-offset", -1, "Begin copying from this offset (applies to all partitions)")
        cpCmd.Flags().Int64("end-offset", -1, "Stop copying at this offset (optional, applies to all partitions)")
}