package cmd

import (
        "crypto/rand"
        "fmt"
        "math/big"

        "github.com/spf13/cobra"
        "kaf/config"
        "kaf/internal/kafka"
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

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                metadata, err := client.DumpMetadata()
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                return formatter.Output(metadata)
        },
}

var utilTailCmd = &cobra.Command{
        Use:   "tail <topic>",
        Short: "Tail messages (like tail -f)",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                topicName := args[0]
                fmt.Printf("Tailing messages from topic '%s' (similar to tail -f)...\n", topicName)
                fmt.Printf("Use: kaf consume %s --from-beginning\n", topicName)
                fmt.Println("For full tail functionality, the consume command provides real-time message streaming")
                return nil
        },
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
        utilCmd.AddCommand(utilTailCmd)

        // Add flags
        utilRandomKeyCmd.Flags().Int("length", 8, "Length of random key")
}