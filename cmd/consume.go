package cmd

import (
        "encoding/json"
        "fmt"
        "os"
        "os/signal"
        "syscall"
        "time"

        "github.com/confluentinc/confluent-kafka-go/v2/kafka"
        "github.com/spf13/cobra"
        "kaf/config"
        kafkaClient "kaf/internal/kafka"
)

var consumeCmd = &cobra.Command{
        Use:   "consume <topic>",
        Short: "Consume messages from a topic",
        Args:  cobra.ExactArgs(1),
        ValidArgsFunction: completeTopics,
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                group, _ := cmd.Flags().GetString("group")
                fromBeginning, _ := cmd.Flags().GetBool("from-beginning")
                fromLatest, _ := cmd.Flags().GetBool("from-latest")
                limit, _ := cmd.Flags().GetInt("limit")
                output, _ := cmd.Flags().GetString("output")
                
                // Validate conflicting flags
                if fromBeginning && fromLatest {
                        return fmt.Errorf("cannot use both --from-beginning and --from-latest flags")
                }
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
                if err != nil {
                        return err
                }

                // Generate group ID if not provided
                if group == "" {
                        group = fmt.Sprintf("kaf-consumer-%d", time.Now().Unix())
                }

                // Create consumer with appropriate offset configuration
                var offsetReset string
                if fromBeginning {
                        offsetReset = "earliest"
                } else if fromLatest {
                        offsetReset = "latest"
                } else {
                        offsetReset = "earliest" // default
                }
                
                consumer, err := client.CreateConsumerWithOffset(group, offsetReset)
                if err != nil {
                        return err
                }
                defer consumer.Close()

                // Subscribe to topic
                err = consumer.SubscribeTopics([]string{topicName}, nil)
                if err != nil {
                        return fmt.Errorf("failed to subscribe to topic: %w", err)
                }

                // Setup signal handling
                sigChan := make(chan os.Signal, 1)
                signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

                messageCount := 0
                fmt.Printf("Consuming from topic '%s' (group: %s). Press Ctrl+C to exit.\n", topicName, group)

                for {
                        select {
                        case sig := <-sigChan:
                                fmt.Printf("\nCaught signal %v: cleaning up and terminating\n", sig)
                                // Close consumer first to leave the group
                                consumer.Close()
                                // Wait a moment for the group to become empty
                                time.Sleep(100 * time.Millisecond)
                                // Clean up auto-generated consumer group (only if it was auto-generated)
                                originalGroup, _ := cmd.Flags().GetString("group")
                                if originalGroup == "" { // Only delete if group was auto-generated
                                        if err := client.DeleteConsumerGroup(group); err != nil {
                                                fmt.Printf("Warning: Failed to delete consumer group '%s': %v\n", group, err)
                                        }
                                }
                                return nil
                        default:
                                msg, err := consumer.ReadMessage(100 * time.Millisecond)
                                if err != nil {
                                        // Check if it's just a timeout
                                        if err.(kafka.Error).Code() == kafka.ErrTimedOut {
                                                continue
                                        }
                                        return fmt.Errorf("consumer error: %w", err)
                                }

                                if msg != nil {
                                        if err := printMessage(msg, output); err != nil {
                                                fmt.Printf("Error formatting message: %v\n", err)
                                        }
                                        
                                        messageCount++
                                        if limit > 0 && messageCount >= limit {
                                                fmt.Printf("\nReached limit of %d messages\n", limit)
                                                return nil
                                        }
                                }
                        }
                }
        },
}

func printMessage(msg *kafka.Message, outputFormat string) error {
        switch outputFormat {
        case "json":
                msgData := map[string]interface{}{
                        "topic":     *msg.TopicPartition.Topic,
                        "partition": msg.TopicPartition.Partition,
                        "offset":    msg.TopicPartition.Offset,
                        "key":       string(msg.Key),
                        "value":     string(msg.Value),
                        "timestamp": msg.Timestamp,
                }
                
                encoder := json.NewEncoder(os.Stdout)
                return encoder.Encode(msgData)
                
        case "yaml":
                // Simple YAML-like output
                fmt.Printf("---\n")
                fmt.Printf("topic: %s\n", *msg.TopicPartition.Topic)
                fmt.Printf("partition: %d\n", msg.TopicPartition.Partition)
                fmt.Printf("offset: %d\n", msg.TopicPartition.Offset)
                fmt.Printf("key: %s\n", string(msg.Key))
                fmt.Printf("value: %s\n", string(msg.Value))
                fmt.Printf("timestamp: %s\n", msg.Timestamp.Format(time.RFC3339))
                
        default: // table format
                fmt.Printf("[%s] Partition: %d, Offset: %d, Key: %s, Value: %s\n",
                        msg.Timestamp.Format("15:04:05"),
                        msg.TopicPartition.Partition,
                        msg.TopicPartition.Offset,
                        string(msg.Key),
                        string(msg.Value))
        }
        
        return nil
}

func init() {
        consumeCmd.Flags().String("group", "", "Consumer group ID (auto-generated if not provided)")
        consumeCmd.Flags().Bool("from-beginning", false, "Start from beginning")
        consumeCmd.Flags().Bool("from-latest", false, "Start from latest messages")
        consumeCmd.Flags().Int("limit", 0, "Limit number of messages (0 = unlimited)")
        consumeCmd.Flags().String("output", "table", "Output format (table, json, yaml)")
}