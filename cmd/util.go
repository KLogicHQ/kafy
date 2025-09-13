package cmd

import (
        "crypto/rand"
        "fmt"
        "math/big"
        "strings"

        "github.com/spf13/cobra"
        "kkl/config"
        kafkaClient "kkl/internal/kafka"
)

var utilCmd = &cobra.Command{
        Use:   "util",
        Short: "Utility commands",
        Long:  "Various utility commands for Kafka operations",
}

var utilRandomKeyCmd = &cobra.Command{
        Use:   "random-key",
        Short: "Generate random Kafka key",
        RunE: func(cmd *cobra.Command, args []string) error {
                length, _ := cmd.Flags().GetInt("length")
                key := generateRandomKey(length)
                fmt.Println(key)
                return nil
        },
}

var utilDumpMetadataCmd = &cobra.Command{
        Use:   "dump-metadata",
        Short: "Dump cluster metadata",
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := config.LoadConfig()
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

                formatter := getFormatter()
                
                // For table format, we need to display brokers and topics separately
                if formatter.Format == "table" {
                        return displayMetadataAsTable(metadata, formatter)
                }
                
                return formatter.Output(metadata)
        },
}

func displayMetadataAsTable(metadata interface{}, formatter interface{}) error {
        data := metadata.(map[string]interface{})
        
        fmt.Println("=== BROKERS ===")
        brokers := data["brokers"].([]map[string]interface{})
        if len(brokers) > 0 {
                brokerHeaders := []string{"ID", "Host", "Port"}
                brokerRows := [][]string{}
                
                for _, broker := range brokers {
                        row := []string{
                                fmt.Sprintf("%v", broker["id"]),
                                fmt.Sprintf("%v", broker["host"]),
                                fmt.Sprintf("%v", broker["port"]),
                        }
                        brokerRows = append(brokerRows, row)
                }
                
                getFormatter().OutputTable(brokerHeaders, brokerRows)
        } else {
                fmt.Println("No brokers found")
        }
        
        fmt.Println("\n=== TOPICS ===")
        topics := data["topics"].([]map[string]interface{})
        if len(topics) > 0 {
                topicHeaders := []string{"Topic", "Partition", "Leader", "Replicas", "ISRs"}
                topicRows := [][]string{}
                
                for _, topic := range topics {
                        topicName := fmt.Sprintf("%v", topic["name"])
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
                                topicRows = append(topicRows, row)
                        }
                }
                
                getFormatter().OutputTable(topicHeaders, topicRows)
        } else {
                fmt.Println("No topics found")
        }
        
        return nil
}


func generateRandomKey(length int) string {
        const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
        b := make([]byte, length)
        for i := range b {
                randomInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
                b[i] = charset[randomInt.Int64()]
        }
        return string(b)
}

func init() {
        utilCmd.AddCommand(utilRandomKeyCmd)
        utilCmd.AddCommand(utilDumpMetadataCmd)

        // Add flags
        utilRandomKeyCmd.Flags().Int("length", 8, "Length of random key")
}