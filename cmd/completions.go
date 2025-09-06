package cmd

import (
        "fmt"
        "kaf/config"
        "kaf/internal/kafka"
        
        "github.com/spf13/cobra"
)

// Dynamic completion functions for Kafka resources

// completeTopics provides completion for topic names
func completeTopics(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        cfg, err := config.LoadConfig()
        if err != nil {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        // Check if we have a current context configured
        if cfg.CurrentContext == "" {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        client, err := kafka.NewClient(cfg)
        if err != nil {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        topics, err := client.ListTopics()
        if err != nil {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        var topicNames []string
        for _, topic := range topics {
                topicNames = append(topicNames, topic.Name)
        }

        return topicNames, cobra.ShellCompDirectiveNoFileComp
}

// completeGroups provides completion for consumer group names
func completeGroups(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        cfg, err := config.LoadConfig()
        if err != nil {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        if cfg.CurrentContext == "" {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        client, err := kafka.NewClient(cfg)
        if err != nil {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        groups, err := client.ListConsumerGroups()
        if err != nil {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        var groupNames []string
        for _, group := range groups {
                groupNames = append(groupNames, group.GroupID)
        }
        
        return groupNames, cobra.ShellCompDirectiveNoFileComp
}

// completeClusters provides completion for cluster names
func completeClusters(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        cfg, err := config.LoadConfig()
        if err != nil {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        var clusterNames []string
        for name := range cfg.Clusters {
                clusterNames = append(clusterNames, name)
        }

        return clusterNames, cobra.ShellCompDirectiveNoFileComp
}

// completeBrokerIDs provides completion for broker IDs
func completeBrokerIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        cfg, err := config.LoadConfig()
        if err != nil {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        if cfg.CurrentContext == "" {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        client, err := kafka.NewClient(cfg)
        if err != nil {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        brokers, err := client.ListBrokers()
        if err != nil {
                return nil, cobra.ShellCompDirectiveNoFileComp
        }

        var brokerIDs []string
        for _, broker := range brokers {
                brokerIDs = append(brokerIDs, fmt.Sprintf("%d", broker.ID))
        }

        return brokerIDs, cobra.ShellCompDirectiveNoFileComp
}

// completeOutputFormats provides completion for output formats
func completeOutputFormats(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        return []string{"table", "json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
}