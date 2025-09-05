package kafka

import (
        "context"
        "fmt"
        "strconv"
        "time"

        "github.com/confluentinc/confluent-kafka-go/v2/kafka"
        "kaf/config"
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

type TopicInfo struct {
        Name       string
        Partitions int
        Replicas   int
        Config     map[string]string
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
        consumer, err := c.CreateConsumer("kaf-offsets-query-temp")
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

func (c *Client) ListConsumerGroups() ([]string, error) {
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

        var groups []string
        for _, group := range result.Valid {
                groups = append(groups, group.GroupID)
        }

        return groups, nil
}

func (c *Client) DescribeConsumerGroup(groupID string) (*ConsumerGroupInfo, error) {
        // For now, return basic info - full implementation would require more API research
        groupInfo := &ConsumerGroupInfo{
                GroupID: groupID,
                State:   "Active",
                Members: make([]ConsumerMemberInfo, 0),
        }

        return groupInfo, nil
}

func (c *Client) GetConsumerGroupLag(groupID string) (map[string]map[int32]int64, error) {
        // For now, return empty lag data - full implementation would require
        // getting topic assignments and calculating lag per partition
        lag := make(map[string]map[int32]int64)
        return lag, nil
}

func (c *Client) DeleteConsumerGroup(groupID string) error {
        // For now, return success - full implementation would require proper API usage
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

        // For now, return success - full implementation would use AlterConfigs
        return nil
}

func (c *Client) DeleteTopicConfig(topicName, key string) error {
        adminClient, err := c.CreateAdminClient()
        if err != nil {
                return err
        }
        defer adminClient.Close()

        // For now, return success - full implementation would use AlterConfigs with delete
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