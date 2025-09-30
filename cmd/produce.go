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
        kafkaClient "kafy/internal/kafka"
)

var produceCmd = &cobra.Command{
        Use:          "produce <topic>",
        Short:        "Produce messages to a topic",
        Args:         cobra.ExactArgs(1),
        ValidArgsFunction: completeTopics,
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                key, _ := cmd.Flags().GetString("key")
                file, _ := cmd.Flags().GetString("file")
                format, _ := cmd.Flags().GetString("format")
                count, _ := cmd.Flags().GetInt("count")
                size, _ := cmd.Flags().GetInt("size")
                headerStrings, _ := cmd.Flags().GetStringSlice("header")

                // Parse headers from command line flags
                headers, err := parseHeaders(headerStrings)
                if err != nil {
                        return fmt.Errorf("invalid header format: %w", err)
                }

                cfg, err := LoadConfigWithClusterOverride()
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
                        return produceTestMessages(producer, topicName, key, headers, count, size)
                }

                if file != "" {
                        return produceFromFile(producer, topicName, key, headers, file, format)
                }

                return produceInteractive(producer, topicName, key, headers, format)
        },
}

// parseHeaders parses header strings in the format "key:value" or "key=value"
func parseHeaders(headerStrings []string) ([]kafka.Header, error) {
        if len(headerStrings) == 0 {
                return nil, nil
        }

        headers := make([]kafka.Header, len(headerStrings))
        for i, headerStr := range headerStrings {
                // Support both "key:value" and "key=value" formats
                var key, value string
                if strings.Contains(headerStr, ":") {
                        parts := strings.SplitN(headerStr, ":", 2)
                        if len(parts) != 2 {
                                return nil, fmt.Errorf("header '%s' must be in format 'key:value' or 'key=value'", headerStr)
                        }
                        key, value = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
                } else if strings.Contains(headerStr, "=") {
                        parts := strings.SplitN(headerStr, "=", 2)
                        if len(parts) != 2 {
                                return nil, fmt.Errorf("header '%s' must be in format 'key:value' or 'key=value'", headerStr)
                        }
                        key, value = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
                } else {
                        return nil, fmt.Errorf("header '%s' must be in format 'key:value' or 'key=value'", headerStr)
                }

                if key == "" {
                        return nil, fmt.Errorf("header key cannot be empty in '%s'", headerStr)
                }

                headers[i] = kafka.Header{
                        Key:   key,
                        Value: []byte(value),
                }
        }

        return headers, nil
}

func produceTestMessages(producer *kafka.Producer, topic, key string, headers []kafka.Header, count int, size int) error {
        for i := 0; i < count; i++ {
                var message string
                
                if size > 0 {
                        // Generate message of specific size
                        message = generateMessageOfSize(i, size)
                } else {
                        // Default message format
                        message = fmt.Sprintf(`{"id": %d, "message": "test message %d", "timestamp": "%s"}`, 
                                i, i, time.Now().Format(time.RFC3339))
                }
                
                messageKey := key
                if messageKey == "" {
                        messageKey = fmt.Sprintf("key-%d", i)
                }

                err := producer.Produce(&kafka.Message{
                        TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
                        Key:            []byte(messageKey),
                        Value:          []byte(message),
                        Headers:        headers,
                }, nil)
                
                if err != nil {
                        return fmt.Errorf("failed to produce message %d: %w", i, err)
                }
        }

        // Wait for all messages to be delivered
        producer.Flush(15 * 1000)
        if size > 0 {
                fmt.Printf("Produced %d test messages of %d bytes each\n", count, size)
        } else {
                fmt.Printf("Produced %d test messages\n", count)
        }
        return nil
}

func generateMessageOfSize(id int, targetSize int) string {
        // Start with basic JSON structure
        baseMessage := fmt.Sprintf(`{"id": %d, "timestamp": "%s", "data": "`, 
                id, time.Now().Format(time.RFC3339))
        suffix := `"}`
        
        // Calculate how much padding we need
        currentSize := len(baseMessage) + len(suffix)
        if currentSize >= targetSize {
                // If base message is already too big, truncate timestamp and use minimal structure
                minMessage := fmt.Sprintf(`{"id":%d,"data":"`, id)
                minSuffix := `"}`
                minSize := len(minMessage) + len(minSuffix)
                
                if minSize >= targetSize {
                        // If even minimal structure is too big, just return what fits
                        if targetSize < 10 {
                                return strings.Repeat("x", targetSize)
                        }
                        return minMessage[:targetSize-len(minSuffix)] + minSuffix
                }
                
                paddingNeeded := targetSize - minSize
                padding := strings.Repeat("x", paddingNeeded)
                return minMessage + padding + minSuffix
        }
        
        // Generate padding to reach target size
        paddingNeeded := targetSize - currentSize
        padding := strings.Repeat("x", paddingNeeded)
        
        return baseMessage + padding + suffix
}

func produceFromFile(producer *kafka.Producer, topic, key string, headers []kafka.Header, filename, format string) error {
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
                        Headers:        headers,
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

func produceInteractive(producer *kafka.Producer, topic, key string, headers []kafka.Header, format string) error {
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
                        Headers:        headers,
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
        produceCmd.Flags().StringSlice("header", []string{}, "Message headers in format 'key:value' or 'key=value' (can be specified multiple times)")
        produceCmd.Flags().String("file", "", "Produce messages from file")
        produceCmd.Flags().String("format", "text", "Message format (text, json, yaml)")
        produceCmd.Flags().Int("count", 0, "Send random test messages")
        produceCmd.Flags().Int("size", 0, "Size in bytes for each generated test message (used with --count)")
}