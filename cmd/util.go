package cmd

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
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
		return fmt.Errorf("dump-metadata command not yet implemented")
	},
}

var utilTailCmd = &cobra.Command{
	Use:   "tail <topic>",
	Short: "Tail messages (like tail -f)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// This is essentially the same as consume with some different defaults
		return fmt.Errorf("tail command not yet implemented - use 'kaf consume %s' instead", args[0])
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