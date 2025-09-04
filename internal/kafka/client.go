package kafka

import (
        "context"
        "fmt"
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