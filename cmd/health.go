package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"kaf/config"
	"kaf/internal/kafka"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Health checks and monitoring",
	Long:  "Commands for checking Kafka cluster health",
}

var healthCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Run all health checks",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running health checks...")
		
		if err := checkBrokers(); err != nil {
			fmt.Printf("❌ Brokers: %v\n", err)
		} else {
			fmt.Println("✅ Brokers: OK")
		}

		return nil
	},
}

var healthBrokersCmd = &cobra.Command{
	Use:   "brokers",
	Short: "Check broker connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		return checkBrokers()
	},
}

func checkBrokers() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client, err := kafka.NewClient(cfg)
	if err != nil {
		return err
	}

	brokers, err := client.ListBrokers()
	if err != nil {
		return fmt.Errorf("failed to connect to brokers: %w", err)
	}

	if len(brokers) == 0 {
		return fmt.Errorf("no brokers found")
	}

	fmt.Printf("Found %d broker(s)\n", len(brokers))
	return nil
}

func init() {
	healthCmd.AddCommand(healthCheckCmd)
	healthCmd.AddCommand(healthBrokersCmd)
}