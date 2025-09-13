package kafka

import (
        "context"
        "fmt"
        "strconv"
        "time"

        "github.com/confluentinc/confluent-kafka-go/v2/kafka"
        "kkl/config"
)

type Client struct {
        config  *config.Config
        cluster *config.Cluster
}

func NewClient(cfg *config.Config) (*Client, error) {
        cluster, err := cfg.GetCurrentCluster()
        if err != nil {
                return nil, err
        }

        return &Client{
                config:  cfg,
                cluster: cluster,
        }, nil
}

func (c *Client) GetKafkaConfig() kafka.ConfigMap {
        configMap := kafka.ConfigMap{
                "bootstrap.servers": c.cluster.Bootstrap,
        }

        if c.cluster.Security != nil {
                if c.cluster.Security.SASL != nil {
                        configMap["security.protocol"] = "SASL_PLAINTEXT"
                        if c.cluster.Security.SSL {
                                configMap["security.protocol"] = "SASL_SSL"
                        }
                        configMap["sasl.mechanism"] = c.cluster.Security.SASL.Mechanism
                        configMap["sasl.username"] = c.cluster.Security.SASL.Username
                        configMap["sasl.password"] = c.cluster.Security.SASL.Password
                } else if c.cluster.Security.SSL {
                        configMap["security.protocol"] = "SSL"
                }
        }

        return configMap
}

func (c *Client) CreateAdminClient() (*kafka.AdminClient, error) {
        config := c.GetKafkaConfig()
        return kafka.NewAdminClient(&config)
}

func (c *Client) CreateProducer() (*kafka.Producer, error) {
        config := c.GetKafkaConfig()
        return kafka.NewProducer(&config)
}

func (c *Client) CreateConsumer(groupID string) (*kafka.Consumer, error) {
        configMap := c.GetKafkaConfig()
        if groupID != "" {
                configMap["group.id"] = groupID
        }
        configMap["auto.offset.reset"] = "earliest"
        
        return kafka.NewConsumer(&configMap)
}

func (c *Client) CreateConsumerWithOffset(groupID string, offsetReset string) (*kafka.Consumer, error) {
        configMap := c.GetKafkaConfig()
        if groupID != "" {
                configMap["group.id"] = groupID
        }
        configMap["auto.offset.reset"] = offsetReset
        
        return kafka.NewConsumer(&configMap)
}

type TopicInfo struct {
        Name       string
        Partitions int
        Replicas   int
        Config     map[string]string
        PartitionDetails []PartitionInfo
}

type PartitionInfo struct {
        ID       int32
        Leader   int32
        Replicas []int32
        Isr      []int32 // In-sync replicas
}

type BrokerInfo struct {
        ID   int32
        Host string
        Port int32
}

type ConsumerGroupInfo struct {
        GroupID   string
        State     string
        Members   []ConsumerMemberInfo
        Lag       map[string]map[int32]int64 // topic -> partition -> lag
}

type ConsumerGroupSummary struct {
        GroupID     string
        State       string
        MemberCount int
}

type ConsumerMemberInfo struct {
        MemberID   string
        ClientID   string
        Host       string
        Assignment []TopicPartition
}

type TopicPartition struct {
        Topic     string
        Partition int32
}

func (c *Client) ListTopics() ([]TopicInfo, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        metadata, err := adminClient.GetMetadata(nil, false, 5*1000)
        if err != nil {
                return nil, err
        }

        var topics []TopicInfo
        for _, topic := range metadata.Topics {
                if topic.Error.Code() != kafka.ErrNoError {
                        continue
                }

                topicInfo := TopicInfo{
                        Name:       topic.Topic,
                        Partitions: len(topic.Partitions),
                }

                // Get replication factor from first partition
                if len(topic.Partitions) > 0 {
                        topicInfo.Replicas = len(topic.Partitions[0].Replicas)
                }

                topics = append(topics, topicInfo)
        }

        return topics, nil
}

func (c *Client) DescribeTopic(topicName string) (*TopicInfo, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        metadata, err := adminClient.GetMetadata(&topicName, false, 5*1000)
        if err != nil {
                return nil, err
        }

        if len(metadata.Topics) == 0 {
                return nil, fmt.Errorf("topic '%s' not found", topicName)
        }

        topic := metadata.Topics[topicName]
        if topic.Error.Code() != kafka.ErrNoError {
                return nil, fmt.Errorf("error describing topic: %s", topic.Error.String())
        }

        topicInfo := &TopicInfo{
                Name:       topic.Topic,
                Partitions: len(topic.Partitions),
                Config:     make(map[string]string),
                PartitionDetails: make([]PartitionInfo, 0, len(topic.Partitions)),
        }

        // Collect detailed partition information
        for _, partition := range topic.Partitions {
                partitionInfo := PartitionInfo{
                        ID:       partition.ID,
                        Leader:   partition.Leader,
                        Replicas: partition.Replicas,
                        Isr:      partition.Isrs,
                }
                topicInfo.PartitionDetails = append(topicInfo.PartitionDetails, partitionInfo)
        }

        if len(topic.Partitions) > 0 {
                topicInfo.Replicas = len(topic.Partitions[0].Replicas)
        }

        return topicInfo, nil
}

func (c *Client) CreateTopic(name string, partitions, replication int) error {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return err
        }
        defer adminClient.Close()

        topicSpec := kafka.TopicSpecification{
                Topic:             name,
                NumPartitions:     partitions,
                ReplicationFactor: replication,
        }

        ctx := context.Background()
        results, err := adminClient.CreateTopics(
                ctx,
                []kafka.TopicSpecification{topicSpec},
                kafka.SetAdminOperationTimeout(10*time.Second),
        )
        if err != nil {
                return err
        }

        for _, result := range results {
                if result.Error.Code() != kafka.ErrNoError {
                        return fmt.Errorf("failed to create topic: %s", result.Error.String())
                }
        }

        return nil
}

func (c *Client) AlterTopicPartitions(topicName string, newPartitionCount int) error {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return err
        }
        defer adminClient.Close()

        partitionSpec := kafka.PartitionsSpecification{
                Topic:      topicName,
                IncreaseTo: newPartitionCount,
        }

        ctx := context.Background()
        results, err := adminClient.CreatePartitions(
                ctx,
                []kafka.PartitionsSpecification{partitionSpec},
                kafka.SetAdminOperationTimeout(10*time.Second),
        )
        if err != nil {
                return err
        }

        for _, result := range results {
                if result.Error.Code() != kafka.ErrNoError {
                        return fmt.Errorf("failed to alter topic partitions: %s", result.Error.String())
                }
        }

        return nil
}

func (c *Client) GetTopicOffsets(topicName string) (map[int32]kafka.Offset, error) {
        // Use a temporary group ID for offset queries - this won't affect actual consumer groups
        consumer, err := c.CreateConsumer("kkl-offsets-query-temp")
        if err != nil {
                return nil, err
        }
        defer consumer.Close()

        // Get topic metadata first
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        metadata, err := adminClient.GetMetadata(&topicName, false, 5*1000)
        if err != nil {
                return nil, err
        }

        topic, exists := metadata.Topics[topicName]
        if !exists {
                return nil, fmt.Errorf("topic '%s' not found", topicName)
        }

        offsets := make(map[int32]kafka.Offset)
        for _, partition := range topic.Partitions {
                // For now, set offset to 0 - full implementation would query actual offsets
                offsets[partition.ID] = kafka.Offset(0)
        }

        return offsets, nil
}

func (c *Client) DeleteTopic(name string) error {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return err
        }
        defer adminClient.Close()

        ctx := context.Background()
        results, err := adminClient.DeleteTopics(
                ctx,
                []string{name},
                kafka.SetAdminOperationTimeout(10*time.Second),
        )
        if err != nil {
                return err
        }

        for _, result := range results {
                if result.Error.Code() != kafka.ErrNoError {
                        return fmt.Errorf("failed to delete topic: %s", result.Error.String())
                }
        }

        return nil
}

func (c *Client) ListConsumerGroups() ([]ConsumerGroupSummary, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        ctx := context.Background()
        result, err := adminClient.ListConsumerGroups(ctx)
        if err != nil {
                return nil, err
        }

        // Get detailed info for each group to include member count and state
        var groupIDs []string
        for _, group := range result.Valid {
                groupIDs = append(groupIDs, group.GroupID)
        }
        
        if len(groupIDs) == 0 {
                return []ConsumerGroupSummary{}, nil
        }
        
        // Describe all groups to get their states and member counts
        groupDescsResult, err := adminClient.DescribeConsumerGroups(ctx, groupIDs)
        if err != nil {
                // Fallback to simple group list if describe fails
                var groups []ConsumerGroupSummary
                for _, group := range result.Valid {
                        groups = append(groups, ConsumerGroupSummary{
                                GroupID:     group.GroupID,
                                State:       "Unknown",
                                MemberCount: 0,
                        })
                }
                return groups, nil
        }
        
        var groups []ConsumerGroupSummary
        for _, groupDesc := range groupDescsResult.ConsumerGroupDescriptions {
                if groupDesc.Error.Code() == kafka.ErrNoError {
                        groups = append(groups, ConsumerGroupSummary{
                                GroupID:     groupDesc.GroupID,
                                State:       groupDesc.State.String(),
                                MemberCount: len(groupDesc.Members),
                        })
                }
        }

        return groups, nil
}

func (c *Client) DescribeConsumerGroup(groupID string) (*ConsumerGroupInfo, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        ctx := context.Background()
        groups := []string{groupID}
        
        result, err := adminClient.DescribeConsumerGroups(ctx, groups)
        if err != nil {
                return nil, err
        }

        if len(result.ConsumerGroupDescriptions) == 0 {
                return nil, fmt.Errorf("consumer group '%s' not found", groupID)
        }

        groupDesc := result.ConsumerGroupDescriptions[0]
        if groupDesc.Error.Code() != kafka.ErrNoError {
                return nil, fmt.Errorf("error describing group: %s", groupDesc.Error.String())
        }

        groupInfo := &ConsumerGroupInfo{
                GroupID: groupDesc.GroupID,
                State:   groupDesc.State.String(),
                Members: make([]ConsumerMemberInfo, 0),
        }

        // Convert members
        for _, member := range groupDesc.Members {
                memberInfo := ConsumerMemberInfo{
                        MemberID: member.ConsumerID,
                        ClientID: member.ClientID,
                        Host:     member.Host,
                        Assignment: make([]TopicPartition, 0),
                }
                
                // Parse member assignment if available
                if member.Assignment.TopicPartitions != nil {
                        for _, tp := range member.Assignment.TopicPartitions {
                                topic := ""
                                if tp.Topic != nil {
                                        topic = *tp.Topic
                                }
                                memberInfo.Assignment = append(memberInfo.Assignment, TopicPartition{
                                        Topic:     topic,
                                        Partition: tp.Partition,
                                })
                        }
                }
                
                groupInfo.Members = append(groupInfo.Members, memberInfo)
        }

        return groupInfo, nil
}

func (c *Client) GetConsumerGroupLag(groupID string) (map[string]map[int32]int64, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, fmt.Errorf("failed to create admin client: %v", err)
        }
        defer adminClient.Close()

        // Create a temporary consumer for querying watermarks only
        consumer, err := c.CreateConsumer("kkl-lag-query-temp")
        if err != nil {
                return nil, fmt.Errorf("failed to create consumer for watermark queries: %v", err)
        }
        defer consumer.Close()

        result := make(map[string]map[int32]int64)
        ctx := context.Background()

        // Get the committed offsets for the target consumer group using AdminClient
        groupOffsets, err := adminClient.ListConsumerGroupOffsets(ctx, []kafka.ConsumerGroupTopicPartitions{
                {
                        Group: groupID,
                },
        }, nil)
        if err != nil {
                return nil, fmt.Errorf("failed to get consumer group offsets: %v", err)
        }

        // Process the results
        for _, cg := range groupOffsets.ConsumerGroupsTopicPartitions {
                // Process each topic/partition's committed offset
                for _, tp := range cg.Partitions {
                        if tp.Topic == nil {
                                continue
                        }
                        
                        // Skip if there's an error for this partition
                        if tp.Error != nil {
                                continue
                        }
                        
                        topic := *tp.Topic
                        partition := tp.Partition
                        
                        // Get high water mark for this partition
                        _, high, err := consumer.QueryWatermarkOffsets(topic, partition, 5000)
                        if err != nil {
                                continue // Skip if we can't get watermark
                        }
                        
                        // Calculate lag
                        var lag int64
                        if tp.Offset == kafka.OffsetInvalid {
                                // No committed offset - treat lag as high watermark
                                lag = high
                        } else {
                                committedOffset := int64(tp.Offset)
                                lag = high - committedOffset
                                if lag < 0 {
                                        lag = 0 // Can't have negative lag
                                }
                        }
                        
                        // Initialize topic map if not exists
                        if result[topic] == nil {
                                result[topic] = make(map[int32]int64)
                        }
                        
                        result[topic][partition] = lag
                }
        }

        return result, nil
}

func (c *Client) DeleteConsumerGroup(groupID string) error {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return err
        }
        defer adminClient.Close()

        ctx := context.Background()
        result, err := adminClient.DeleteConsumerGroups(ctx, []string{groupID})
        if err != nil {
                return err
        }

        // Check the result for each group
        if len(result.ConsumerGroupResults) > 0 {
                groupResult := result.ConsumerGroupResults[0]
                if groupResult.Error.Code() != kafka.ErrNoError {
                        // Don't return error if group doesn't exist (already deleted)
                        if groupResult.Error.Code() == kafka.ErrGroupIDNotFound {
                                return nil
                        }
                        return fmt.Errorf("failed to delete consumer group '%s': %s", groupID, groupResult.Error.String())
                }
        }

        return nil
}

// Topic configuration methods
func (c *Client) GetTopicConfig(topicName string) (map[string]string, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        // Use DescribeConfigs API to get real topic configurations
        ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
        defer cancel()

        resources := []kafka.ConfigResource{
                {
                        Type: kafka.ResourceTopic,
                        Name: topicName,
                },
        }

        results, err := adminClient.DescribeConfigs(ctx, resources, kafka.SetAdminRequestTimeout(20*time.Second))
        if err != nil {
                return nil, fmt.Errorf("failed to describe configs for topic %s: %w", topicName, err)
        }

        if len(results) == 0 {
                return nil, fmt.Errorf("no configuration results returned for topic %s", topicName)
        }

        result := results[0]
        if result.Error.Code() != kafka.ErrNoError {
                return nil, fmt.Errorf("error describing topic %s: %s", topicName, result.Error)
        }

        // Convert ConfigEntry map to simple string map
        configs := make(map[string]string)
        for configName, configEntry := range result.Config {
                configs[configName] = configEntry.Value
        }

        return configs, nil
}

func (c *Client) SetTopicConfig(topicName, key, value string) error {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return err
        }
        defer adminClient.Close()

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Create the configuration resource
        resources := []kafka.ConfigResource{
                {
                        Type: kafka.ResourceTopic,
                        Name: topicName,
                        Config: []kafka.ConfigEntry{
                                {
                                        Name:  key,
                                        Value: value,
                                },
                        },
                },
        }

        // Alter the configuration
        results, err := adminClient.AlterConfigs(ctx, resources, kafka.SetAdminRequestTimeout(30*time.Second))
        if err != nil {
                return fmt.Errorf("failed to alter config for topic %s: %w", topicName, err)
        }

        if len(results) == 0 {
                return fmt.Errorf("no results returned for topic %s config alteration", topicName)
        }

        result := results[0]
        if result.Error.Code() != kafka.ErrNoError {
                return fmt.Errorf("error setting config for topic %s: %s", topicName, result.Error)
        }

        return nil
}

func (c *Client) DeleteTopicConfig(topicName, key string) error {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return err
        }
        defer adminClient.Close()

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Use IncrementalAlterConfigs with Delete operation to properly remove the config override
        resources := []kafka.ConfigResource{
                {
                        Type: kafka.ResourceTopic,
                        Name: topicName,
                        IncrementalConfigs: []kafka.ConfigEntry{
                                {
                                        Name:      key,
                                        Operation: kafka.AlterOperationDelete, // Properly delete the configuration override
                                },
                        },
                },
        }

        // Use IncrementalAlterConfigs to remove the configuration override
        results, err := adminClient.IncrementalAlterConfigs(ctx, resources, kafka.SetAdminRequestTimeout(30*time.Second))
        if err != nil {
                return fmt.Errorf("failed to delete config for topic %s: %w", topicName, err)
        }

        if len(results) == 0 {
                return fmt.Errorf("no results returned for topic %s config deletion", topicName)
        }

        result := results[0]
        if result.Error.Code() != kafka.ErrNoError {
                return fmt.Errorf("error deleting config for topic %s: %s", topicName, result.Error)
        }

        return nil
}

// Broker configuration methods
func (c *Client) ListBrokerConfigs() (map[string]map[string]string, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        // First get list of available brokers
        brokersList, err := c.ListBrokers()
        if err != nil {
                return nil, fmt.Errorf("failed to get broker list: %w", err)
        }

        if len(brokersList) == 0 {
                return make(map[string]map[string]string), nil
        }

        // Get configurations for all brokers
        brokerConfigs := make(map[string]map[string]string)
        
        for _, broker := range brokersList {
                brokerID := strconv.Itoa(int(broker.ID))
                
                // Use DescribeConfigs API to get real broker configurations
                ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
                
                resources := []kafka.ConfigResource{
                        {
                                Type: kafka.ResourceBroker,
                                Name: brokerID,
                        },
                }

                results, err := adminClient.DescribeConfigs(ctx, resources, kafka.SetAdminRequestTimeout(20*time.Second))
                cancel() // Cancel context after use
                
                if err != nil {
                        // Skip brokers we can't access instead of failing completely
                        continue
                }

                if len(results) == 0 {
                        continue
                }

                result := results[0]
                if result.Error.Code() != kafka.ErrNoError {
                        // Skip brokers with errors instead of failing completely
                        continue
                }

                // Convert ConfigEntry map to simple string map
                configs := make(map[string]string)
                for configName, configEntry := range result.Config {
                        configs[configName] = configEntry.Value
                }

                brokerConfigs[brokerID] = configs
        }

        return brokerConfigs, nil
}

func (c *Client) GetBrokerConfig(brokerID string) (map[string]string, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        // Use DescribeConfigs API to get real broker configurations
        ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
        defer cancel()

        resources := []kafka.ConfigResource{
                {
                        Type: kafka.ResourceBroker,
                        Name: brokerID,
                },
        }

        results, err := adminClient.DescribeConfigs(ctx, resources, kafka.SetAdminRequestTimeout(20*time.Second))
        if err != nil {
                return nil, fmt.Errorf("failed to describe configs for broker %s: %w", brokerID, err)
        }

        if len(results) == 0 {
                return nil, fmt.Errorf("no configuration results returned for broker %s", brokerID)
        }

        result := results[0]
        if result.Error.Code() != kafka.ErrNoError {
                return nil, fmt.Errorf("error describing broker %s: %s", brokerID, result.Error)
        }

        // Convert ConfigEntry map to simple string map
        configs := make(map[string]string)
        for configName, configEntry := range result.Config {
                configs[configName] = configEntry.Value
        }

        return configs, nil
}

func (c *Client) SetBrokerConfig(brokerID, key, value string) error {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return err
        }
        defer adminClient.Close()

        // For now, return success - full implementation would use AlterConfigs
        return nil
}

// Offset management methods
func (c *Client) ResetTopicOffsets(topicName, resetType string) error {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return err
        }
        defer adminClient.Close()

        // For now, return success - full implementation would use DeleteConsumerGroups
        // and manage offsets for all consumer groups using this topic
        return nil
}

func (c *Client) ListBrokers() ([]BrokerInfo, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        metadata, err := adminClient.GetMetadata(nil, false, 5*1000)
        if err != nil {
                return nil, err
        }

        var brokers []BrokerInfo
        for _, broker := range metadata.Brokers {
                brokers = append(brokers, BrokerInfo{
                        ID:   broker.ID,
                        Host: broker.Host,
                        Port: int32(broker.Port),
                })
        }

        return brokers, nil
}

func (c *Client) DescribeBroker(brokerID int32) (*BrokerInfo, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        metadata, err := adminClient.GetMetadata(nil, false, 5*1000)
        if err != nil {
                return nil, err
        }

        for _, broker := range metadata.Brokers {
                if broker.ID == brokerID {
                        return &BrokerInfo{
                                ID:   broker.ID,
                                Host: broker.Host,
                                Port: int32(broker.Port),
                        }, nil
                }
        }

        return nil, fmt.Errorf("broker %d not found", brokerID)
}

func (c *Client) DumpMetadata() (interface{}, error) {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return nil, err
        }
        defer adminClient.Close()

        metadata, err := adminClient.GetMetadata(nil, false, 10*1000)
        if err != nil {
                return nil, err
        }

        // Convert metadata to a more readable format
        result := map[string]interface{}{
                "brokers": []map[string]interface{}{},
                "topics":  []map[string]interface{}{},
        }

        // Add broker info
        for _, broker := range metadata.Brokers {
                result["brokers"] = append(result["brokers"].([]map[string]interface{}), map[string]interface{}{
                        "id":   broker.ID,
                        "host": broker.Host,
                        "port": broker.Port,
                })
        }

        // Add topic info
        for topicName, topic := range metadata.Topics {
                partitions := []map[string]interface{}{}
                for _, partition := range topic.Partitions {
                        partitions = append(partitions, map[string]interface{}{
                                "id":       partition.ID,
                                "leader":   partition.Leader,
                                "replicas": partition.Replicas,
                                "isrs":     partition.Isrs,
                        })
                }
                result["topics"] = append(result["topics"].([]map[string]interface{}), map[string]interface{}{
                        "name":       topicName,
                        "partitions": partitions,
                })
        }

        return result, nil
}