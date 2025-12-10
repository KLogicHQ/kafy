package cmd

import (
        "fmt"
        "os"
        "os/signal"
        "strings"
        "syscall"
        "time"

        "github.com/confluentinc/confluent-kafka-go/v2/kafka"
        "github.com/spf13/cobra"
        kafkaClient "kafy/internal/kafka"
)

var tailCmd = &cobra.Command{
        Use:   "tail <topic1> [topic2] [topic3] ...",
        Short: "Tail messages in real-time from one or more topics (like tail -f)",
        Args:  cobra.MinimumNArgs(1),
        ValidArgsFunction: completeTopics,
        RunE: func(cmd *cobra.Command, args []string) error {
                topicNames := args
                output, _ := cmd.Flags().GetString("output")
                keyFilter, _ := cmd.Flags().GetString("key-filter")
                noValue, _ := cmd.Flags().GetBool("no-value")

                // Execute consume with --from-latest flag
                cfg, err := LoadConfigWithClusterOverride()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
                if err != nil {
                        return err
                }

                // Generate unique group ID for tail
                group := fmt.Sprintf("kafy-tail-%d", time.Now().Unix())

                consumer, err := client.CreateConsumerWithOffset(group, "latest")
                if err != nil {
                        return err
                }
                defer consumer.Close()

                // Subscribe to topics
                err = consumer.SubscribeTopics(topicNames, nil)
                if err != nil {
                        return fmt.Errorf("failed to subscribe to topics: %w", err)
                }

                // Setup signal handling
                sigChan := make(chan os.Signal, 1)
                signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

                if len(topicNames) == 1 {
                        fmt.Printf("Tailing messages from topic '%s' (latest messages only). Press Ctrl+C to exit.\n", topicNames[0])
                } else {
                        fmt.Printf("Tailing messages from %d topics: %s (latest messages only). Press Ctrl+C to exit.\n", len(topicNames), strings.Join(topicNames, ", "))
                }

                for {
                        select {
                        case sig := <-sigChan:
                                fmt.Printf("\nCaught signal %v: cleaning up and terminating\n", sig)
                                // Close consumer first to leave the group
                                consumer.Close()
                                // Wait a moment for the group to become empty
                                time.Sleep(100 * time.Millisecond)
                                // Clean up the auto-generated consumer group
                                if err := client.DeleteConsumerGroup(group); err != nil {
                                        fmt.Printf("Warning: Failed to delete consumer group '%s': %v\n", group, err)
                                }
                                return nil
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

                                // Apply key filtering if specified
                                if keyFilter != "" {
                                        messageKey := string(msg.Key)
                                        if !matchesKeyFilter(messageKey, keyFilter) {
                                                continue // Skip this message
                                        }
                                }
                                
                                // Use the same message printing function as consume command
                                if err := printMessage(msg, output, noValue); err != nil {
                                        fmt.Printf("Error formatting message: %v\n", err)
                                }
                        }
                }
        },
}

func init() {
        tailCmd.Flags().String("output", "table", "Output format (table, json, yaml, hex)")
        tailCmd.Flags().String("key-filter", "", "Filter messages by key (supports wildcards: *, prefix*, *suffix, *contains*)")
        tailCmd.Flags().Bool("no-value", false, "Hide message values from output")
}