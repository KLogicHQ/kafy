package cmd

import (
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

var tailCmd = &cobra.Command{
	Use:   "tail <topic>",
	Short: "Tail messages in real-time (like tail -f)",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: completeTopics,
	RunE: func(cmd *cobra.Command, args []string) error {
		topicName := args[0]
		
		// Execute consume with --from-latest flag
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client, err := kafkaClient.NewClient(cfg)
		if err != nil {
			return err
		}

		// Generate unique group ID for tail
		group := fmt.Sprintf("kaf-tail-%d", time.Now().Unix())

		consumer, err := client.CreateConsumerWithOffset(group, "latest")
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

		fmt.Printf("Tailing messages from topic '%s' (latest messages only). Press Ctrl+C to exit.\n", topicName)

		for {
			select {
			case sig := <-sigChan:
				fmt.Printf("\nCaught signal %v: terminating\n", sig)
				return nil
			default:
				msg, err := consumer.ReadMessage(100 * time.Millisecond)
				if err != nil {
					// Check if it's just a timeout
					if err.(kafka.Error).Code() == kafka.ErrTimedOut {
						continue
					}
					return fmt.Errorf("failed to read message: %w", err)
				}

				// Display message in table format (similar to consume command)
				fmt.Printf("[%s] Partition: %d, Offset: %d, Key: %s, Value: %s\n",
					msg.Timestamp.Format("15:04:05"),
					msg.TopicPartition.Partition,
					msg.TopicPartition.Offset,
					string(msg.Key),
					string(msg.Value))
			}
		}
	},
}