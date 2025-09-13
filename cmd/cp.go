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
  kkl cp --limit 1000 transactions transactions-test`,
        Args:              cobra.ExactArgs(2),
        ValidArgsFunction: completeTopics,
        RunE: func(cmd *cobra.Command, args []string) error {
                sourceTopic := args[0]
                destTopic := args[1]
                fromBeginning, _ := cmd.Flags().GetBool("from-beginning")
                limit, _ := cmd.Flags().GetInt("limit")
                
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

                // Subscribe to source topic
                err = consumer.SubscribeTopics([]string{sourceTopic}, nil)
                if err != nil {
                        return fmt.Errorf("failed to subscribe to source topic: %w", err)
                }

                fmt.Printf("Starting to copy messages from '%s' to '%s'...\n", sourceTopic, destTopic)
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
}