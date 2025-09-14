package cmd

import (
        "fmt"

        "github.com/spf13/cobra"
        kafkaClient "kafy/internal/kafka"
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
                fmt.Println("Running comprehensive health checks...")
                fmt.Println()
                
                // Check brokers
                fmt.Println("🔍 Checking brokers...")
                if err := checkBrokers(); err != nil {
                        fmt.Printf("❌ Brokers: %v\n", err)
                } else {
                        fmt.Println("✅ Brokers: OK")
                }
                fmt.Println()

                // Check topics
                fmt.Println("🔍 Checking topics...")
                if err := checkTopics(); err != nil {
                        fmt.Printf("❌ Topics: %v\n", err)
                } else {
                        fmt.Println("✅ Topics: OK")
                }
                fmt.Println()

                // Check groups
                fmt.Println("🔍 Checking consumer groups...")
                if err := checkGroups(); err != nil {
                        fmt.Printf("❌ Consumer Groups: %v\n", err)
                } else {
                        fmt.Println("✅ Consumer Groups: OK")
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
        cfg, err := LoadConfigWithClusterOverride()
        if err != nil {
                return err
        }

        client, err := kafkaClient.NewClient(cfg)
        if err != nil {
                return err
        }

        // Use GetMetadata to get comprehensive broker information including unavailable brokers
        adminClient, err := client.CreateAdminClient()
        if err != nil {
                return fmt.Errorf("failed to create admin client: %w", err)
        }
        defer adminClient.Close()

        metadata, err := adminClient.GetMetadata(nil, false, 5*1000)
        if err != nil {
                return fmt.Errorf("failed to get cluster metadata: %w", err)
        }

        if len(metadata.Brokers) == 0 {
                return fmt.Errorf("no brokers found in cluster metadata")
        }

        fmt.Printf("Found %d broker(s) in cluster metadata:\n", len(metadata.Brokers))
        
        live := 0
        for _, broker := range metadata.Brokers {
                // Try to get additional broker information to test connectivity
                liveBrokers, err := client.ListBrokers()
                brokerLive := false
                if err == nil {
                        for _, liveBroker := range liveBrokers {
                                if liveBroker.ID == broker.ID {
                                        brokerLive = true
                                        break
                                }
                        }
                }
                
                if brokerLive {
                        fmt.Printf("  ✅ Broker %d (%s:%d) - Live\n", broker.ID, broker.Host, broker.Port)
                        live++
                } else {
                        fmt.Printf("  ❌ Broker %d (%s:%d) - Unreachable\n", broker.ID, broker.Host, broker.Port)
                }
        }

        fmt.Printf("Connectivity: %d/%d brokers are live\n", live, len(metadata.Brokers))
        
        if live == 0 {
                return fmt.Errorf("no brokers are reachable")
        }

        return nil
}

var healthTopicsCmd = &cobra.Command{
        Use:   "topics",
        Short: "Check topic accessibility and health",
        RunE: func(cmd *cobra.Command, args []string) error {
                return checkTopics()
        },
}

func checkTopics() error {
        cfg, err := LoadConfigWithClusterOverride()
        if err != nil {
                return err
        }

        client, err := kafkaClient.NewClient(cfg)
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
}

var healthGroupsCmd = &cobra.Command{
        Use:   "groups",
        Short: "Check consumer group health and status",
        RunE: func(cmd *cobra.Command, args []string) error {
                return checkGroups()
        },
}

func checkGroups() error {
        cfg, err := LoadConfigWithClusterOverride()
        if err != nil {
                return err
        }

        client, err := kafkaClient.NewClient(cfg)
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
                _, err := client.DescribeConsumerGroup(group.GroupID)
                if err != nil {
                        fmt.Printf("  ⚠ Group '%s': %v\n", group.GroupID, err)
                } else {
                        healthy++
                        fmt.Printf("  ✓ Group '%s': accessible\n", group.GroupID)
                }
        }
        
        if len(groups) > 5 {
                fmt.Printf("  ... and %d more groups\n", len(groups)-5)
        }
        
        fmt.Printf("✓ %d/%d sampled groups are healthy\n", healthy, checkCount)
        return nil
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