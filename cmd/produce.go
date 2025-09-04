package cmd

import (
        "bufio"
        "encoding/json"
        "fmt"
        "os"
        "strings"
        "time"

        "github.com/confluentinc/confluent-kafka-go/v2/kafka"
        "github.com/spf13/cobra"
        "kaf/config"
        kafkaClient "kaf/internal/kafka"
)

var produceCmd = &cobra.Command{
        Use:   "produce <topic>",
        Short: "Produce messages to a topic",
        Args:  cobra.ExactArgs(1),
        ValidArgsFunction: completeTopics,
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                key, _ := cmd.Flags().GetString("key")
                file, _ := cmd.Flags().GetString("file")
                format, _ := cmd.Flags().GetString("format")
                count, _ := cmd.Flags().GetInt("count")
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafkaClient.NewClient(cfg)
                if err != nil {
                        return err
                }

                producer, err := client.CreateProducer()
                if err != nil {
                        return err
                }
                defer producer.Close()

                // Handle delivery reports
                go func() {
                        for e := range producer.Events() {
                                switch ev := e.(type) {
                                case *kafka.Message:
                                        if ev.TopicPartition.Error != nil {
                                                fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
                                        } else {
                                                fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
                                        }
                                }
                        }
                }()

                if count > 0 {
                        return produceTestMessages(producer, topicName, key, count)
                }

                if file != "" {
                        return produceFromFile(producer, topicName, key, file, format)
                }

                return produceInteractive(producer, topicName, key, format)
        },
}

func produceTestMessages(producer *kafka.Producer, topic, key string, count int) error {
        for i := 0; i < count; i++ {
                message := fmt.Sprintf(`{"id": %d, "message": "test message %d", "timestamp": "%s"}`, 
                        i, i, time.Now().Format(time.RFC3339))
                
                messageKey := key
                if messageKey == "" {
                        messageKey = fmt.Sprintf("key-%d", i)
                }

                err := producer.Produce(&kafka.Message{
                        TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
                        Key:            []byte(messageKey),
                        Value:          []byte(message),
                }, nil)
                
                if err != nil {
                        return fmt.Errorf("failed to produce message %d: %w", i, err)
                }
        }

        // Wait for all messages to be delivered
        producer.Flush(15 * 1000)
        fmt.Printf("Produced %d test messages\n", count)
        return nil
}

func produceFromFile(producer *kafka.Producer, topic, key, filename, format string) error {
        file, err := os.Open(filename)
        if err != nil {
                return fmt.Errorf("failed to open file: %w", err)
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)
        messageCount := 0

        for scanner.Scan() {
                line := strings.TrimSpace(scanner.Text())
                if line == "" {
                        continue
                }

                var messageValue []byte
                switch format {
                case "json":
                        // Validate JSON
                        var js json.RawMessage
                        if err := json.Unmarshal([]byte(line), &js); err != nil {
                                fmt.Printf("Skipping invalid JSON: %s\n", line)
                                continue
                        }
                        messageValue = []byte(line)
                case "yaml":
                        // For now, treat as plain text
                        messageValue = []byte(line)
                default:
                        messageValue = []byte(line)
                }

                messageKey := key
                if messageKey == "" {
                        messageKey = fmt.Sprintf("msg-%d", messageCount)
                }

                err := producer.Produce(&kafka.Message{
                        TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
                        Key:            []byte(messageKey),
                        Value:          messageValue,
                }, nil)
                
                if err != nil {
                        return fmt.Errorf("failed to produce message: %w", err)
                }
                
                messageCount++
        }

        if err := scanner.Err(); err != nil {
                return fmt.Errorf("error reading file: %w", err)
        }

        // Wait for all messages to be delivered
        producer.Flush(15 * 1000)
        fmt.Printf("Produced %d messages from file\n", messageCount)
        return nil
}

func produceInteractive(producer *kafka.Producer, topic, key, format string) error {
        fmt.Printf("Producing messages to topic '%s'. Type messages and press Enter. Ctrl+C to exit.\n", topic)
        
        scanner := bufio.NewScanner(os.Stdin)
        messageCount := 0

        for scanner.Scan() {
                line := strings.TrimSpace(scanner.Text())
                if line == "" {
                        continue
                }

                var messageValue []byte
                switch format {
                case "json":
                        // Validate JSON
                        var js json.RawMessage
                        if err := json.Unmarshal([]byte(line), &js); err != nil {
                                fmt.Printf("Invalid JSON, sending as plain text: %s\n", line)
                        }
                        messageValue = []byte(line)
                default:
                        messageValue = []byte(line)
                }

                messageKey := key
                if messageKey == "" {
                        messageKey = fmt.Sprintf("msg-%d", messageCount)
                }

                err := producer.Produce(&kafka.Message{
                        TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
                        Key:            []byte(messageKey),
                        Value:          messageValue,
                }, nil)
                
                if err != nil {
                        fmt.Printf("Failed to produce message: %v\n", err)
                        continue
                }
                
                messageCount++
                fmt.Printf("Message %d queued\n", messageCount)
        }

        if err := scanner.Err(); err != nil {
                return fmt.Errorf("error reading input: %w", err)
        }

        // Wait for all messages to be delivered
        producer.Flush(15 * 1000)
        return nil
}

func init() {
        produceCmd.Flags().String("key", "", "Message key")
        produceCmd.Flags().String("file", "", "Produce messages from file")
        produceCmd.Flags().String("format", "text", "Message format (text, json, yaml)")
        produceCmd.Flags().Int("count", 0, "Send random test messages")
}