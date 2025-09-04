package cmd

import (
        "kaf/config"
        "kaf/internal/kafka"
        
        "github.com/spf13/cobra"
)

// Dynamic completion functions for Kafka resources

// completeTopics provides completion for topic names
func completeTopics(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        cfg, err := config.LoadConfig()
        if err != nil {
                return nil, cobra.ShellCompDirectiveError
        }

        client, err := kafka.NewClient(cfg)
        if err != nil {
                return nil, cobra.ShellCompDirectiveError
        }

        topics, err := client.ListTopics()
        if err != nil {
                return nil, cobra.ShellCompDirectiveError
        }

        var topicNames []string
        for _, topic := range topics {
                topicNames = append(topicNames, topic.Name)
        }

        return topicNames, cobra.ShellCompDirectiveDefault
}

// completeGroups provides completion for consumer group names
func completeGroups(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        cfg, err := config.LoadConfig()
        if err != nil {
                return nil, cobra.ShellCompDirectiveError
        }

        client, err := kafka.NewClient(cfg)
        if err != nil {
                return nil, cobra.ShellCompDirectiveError
        }

        groups, err := client.ListConsumerGroups()
        if err != nil {
                return nil, cobra.ShellCompDirectiveError
        }

        return groups, cobra.ShellCompDirectiveDefault
}

// completeClusters provides completion for cluster names
func completeClusters(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        cfg, err := config.LoadConfig()
        if err != nil {
                return nil, cobra.ShellCompDirectiveError
        }

        var clusterNames []string
        for name := range cfg.Clusters {
                clusterNames = append(clusterNames, name)
        }

        return clusterNames, cobra.ShellCompDirectiveDefault
}

// completeBrokerIDs provides completion for broker IDs
func completeBrokerIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        cfg, err := config.LoadConfig()
        if err != nil {
                return nil, cobra.ShellCompDirectiveError
        }

        client, err := kafka.NewClient(cfg)
        if err != nil {
                return nil, cobra.ShellCompDirectiveError
        }

        brokers, err := client.ListBrokers()
        if err != nil {
                return nil, cobra.ShellCompDirectiveError
        }

        var brokerIDs []string
        for _, broker := range brokers {
                brokerIDs = append(brokerIDs, string(rune(broker.ID + '0')))
        }

        return brokerIDs, cobra.ShellCompDirectiveDefault
}

// completeOutputFormats provides completion for output formats
func completeOutputFormats(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        return []string{"table", "json", "yaml"}, cobra.ShellCompDirectiveDefault
}