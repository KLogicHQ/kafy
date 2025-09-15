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
        
        // Get list of all broker IDs that are expected from topic replicas
        expectedBrokerIDs, err := getExpectedBrokerIDsFromTopics(client)
        if err != nil {
                fmt.Printf("Warning: Could not determine expected brokers from topics: %v\n", err)
                expectedBrokerIDs = make(map[int32]bool) // Empty set, fallback to all brokers from metadata
        }
        
        live := 0
        liveBrokers, err := client.ListBrokers()
        liveSet := make(map[int32]bool)
        if err == nil {
                for _, broker := range liveBrokers {
                        liveSet[broker.ID] = true
                }
        }
        
        for _, broker := range metadata.Brokers {
                brokerLive := liveSet[broker.ID]
                
                if brokerLive {
                        fmt.Printf("  ✅ Broker %d (%s:%d) - Live\n", broker.ID, broker.Host, broker.Port)
                        live++
                } else {
                        fmt.Printf("  ❌ Broker %d (%s:%d) - Unreachable\n", broker.ID, broker.Host, broker.Port)
                }
        }

        fmt.Printf("Connectivity: %d/%d brokers are live\n", live, len(metadata.Brokers))
        
        // Check if any expected brokers from topics are missing
        if len(expectedBrokerIDs) > 0 {
                fmt.Println("\nTopic Replica Health:")
                missingBrokers := []int32{}
                for brokerID := range expectedBrokerIDs {
                        if !liveSet[brokerID] {
                                missingBrokers = append(missingBrokers, brokerID)
                        }
                }
                
                if len(missingBrokers) == 0 {
                        fmt.Printf("  ✅ All brokers required by topic replicas are healthy (%d brokers)\n", len(expectedBrokerIDs))
                } else {
                        fmt.Printf("  ❌ Missing broker IDs required by topics: %v\n", missingBrokers)
                        fmt.Printf("  ⚠️  This may cause partition unavailability\n")
                }
        }
        
        if live == 0 {
                return fmt.Errorf("no brokers are reachable")
        }

        return nil
}

// getExpectedBrokerIDsFromTopics extracts all broker IDs from topic replica arrays
func getExpectedBrokerIDsFromTopics(client *kafkaClient.Client) (map[int32]bool, error) {
        brokerIDs := make(map[int32]bool)
        
        topics, err := client.ListTopics()
        if err != nil {
                return nil, fmt.Errorf("failed to list topics: %w", err)
        }
        
        // For each topic, get its partition metadata and collect broker IDs from replicas
        for _, topic := range topics {
                topicDetails, err := client.DescribeTopic(topic.Name)
                if err != nil {
                        // Skip topics we can't describe, continue with others
                        continue
                }
                
                // Extract broker IDs from all partition replicas
                for _, partition := range topicDetails.PartitionDetails {
                        for _, replicaID := range partition.Replicas {
                                brokerIDs[replicaID] = true
                        }
                }
        }
        
        return brokerIDs, nil
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