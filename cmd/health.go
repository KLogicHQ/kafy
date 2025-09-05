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

var healthTopicsCmd = &cobra.Command{
        Use:   "topics",
        Short: "Check topic accessibility and health",
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                topics, err := client.ListTopics()
                if err != nil {
                        return fmt.Errorf("failed to list topics: %w", err)
                }

                fmt.Printf("✓ Successfully accessed %d topics\n", len(topics))
                
                // Check a few topics for detailed health
                healthy := 0
                checkCount := min(5, len(topics))
                for i := 0; i < checkCount; i++ {
                        topic := topics[i]
                        
                        // Try to get topic details to verify accessibility
                        _, err := client.DescribeTopic(topic.Name)
                        if err != nil {
                                fmt.Printf("  ⚠ Topic '%s': %v\n", topic.Name, err)
                        } else {
                                healthy++
                                fmt.Printf("  ✓ Topic '%s': accessible\n", topic.Name)
                        }
                }
                
                if len(topics) > 5 {
                        fmt.Printf("  ... and %d more topics\n", len(topics)-5)
                }
                
                fmt.Printf("✓ %d/%d sampled topics are healthy\n", healthy, checkCount)
                return nil
        },
}

var healthGroupsCmd = &cobra.Command{
        Use:   "groups",
        Short: "Check consumer group health and status",
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                groups, err := client.ListConsumerGroups()
                if err != nil {
                        return fmt.Errorf("failed to list consumer groups: %w", err)
                }

                fmt.Printf("✓ Found %d consumer groups\n", len(groups))
                
                // Check a few groups for detailed health
                healthy := 0
                checkCount := min(5, len(groups))
                for i := 0; i < checkCount; i++ {
                        group := groups[i]
                        
                        // Try to describe group to verify accessibility
                        _, err := client.DescribeConsumerGroup(group)
                        if err != nil {
                                fmt.Printf("  ⚠ Group '%s': %v\n", group, err)
                        } else {
                                healthy++
                                fmt.Printf("  ✓ Group '%s': accessible\n", group)
                        }
                }
                
                if len(groups) > 5 {
                        fmt.Printf("  ... and %d more groups\n", len(groups)-5)
                }
                
                fmt.Printf("✓ %d/%d sampled groups are healthy\n", healthy, checkCount)
                return nil
        },
}

func min(a, b int) int {
        if a < b {
                return a
        }
        return b
}

func init() {
        healthCmd.AddCommand(healthCheckCmd)
        healthCmd.AddCommand(healthBrokersCmd)
        healthCmd.AddCommand(healthTopicsCmd)
        healthCmd.AddCommand(healthGroupsCmd)
}