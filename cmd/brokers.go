package cmd

import (
        "fmt"
        "strconv"
        "strings"

        "github.com/spf13/cobra"
        "kaf/config"
        "kaf/internal/kafka"
)

var brokersCmd = &cobra.Command{
        Use:   "brokers",
        Short: "Manage Kafka brokers",
        Long:  "Commands for managing and inspecting Kafka brokers",
}

var brokersListCmd = &cobra.Command{
        Use:   "list",
        Short: "List all brokers",
        RunE: func(cmd *cobra.Command, args []string) error {
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
                        return err
                }

                formatter := getFormatter()
                headers := []string{"ID", "Host", "Port"}
                var rows [][]string

                for _, broker := range brokers {
                        rows = append(rows, []string{
                                strconv.Itoa(int(broker.ID)),
                                broker.Host,
                                strconv.Itoa(int(broker.Port)),
                        })
                }

                formatter.OutputTable(headers, rows)
                return nil
        },
}

var brokersDescribeCmd = &cobra.Command{
        Use:   "describe <broker-id>",
        Short: "Show broker metadata",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                brokerIDStr := args[0]
                brokerID, err := strconv.Atoi(brokerIDStr)
                if err != nil {
                        return fmt.Errorf("invalid broker ID: %s", brokerIDStr)
                }
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                broker, err := client.DescribeBroker(int32(brokerID))
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                if formatter.Format == "table" {
                        headers := []string{"Property", "Value"}
                        rows := [][]string{
                                {"ID", strconv.Itoa(int(broker.ID))},
                                {"Host", broker.Host},
                                {"Port", strconv.Itoa(int(broker.Port))},
                        }
                        formatter.OutputTable(headers, rows)
                        return nil
                }
                
                return formatter.Output(broker)
        },
}

var brokersMetricsCmd = &cobra.Command{
        Use:   "metrics",
        Short: "Show broker metrics (disk, CPU, requests)",
        RunE: func(cmd *cobra.Command, args []string) error {
                fmt.Println("Broker metrics functionality would require JMX or other monitoring integration")
                fmt.Println("This feature would show:")
                fmt.Println("- Disk usage per broker")
                fmt.Println("- CPU utilization")
                fmt.Println("- Request rates")
                fmt.Println("- Network I/O statistics")
                return nil
        },
}

// Broker config commands
var brokersConfigsCmd = &cobra.Command{
        Use:   "configs",
        Short: "Manage broker configurations",
        Long:  "Commands for managing broker-level configurations",
}

var brokersConfigsListCmd = &cobra.Command{
        Use:   "list",
        Short: "List all broker configs",
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                brokerConfigs, err := client.ListBrokerConfigs()
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                headers := []string{"Broker ID", "Config Key", "Value"}
                var rows [][]string

                for brokerID, configs := range brokerConfigs {
                        for key, value := range configs {
                                rows = append(rows, []string{brokerID, key, value})
                        }
                }

                formatter.OutputTable(headers, rows)
                return nil
        },
}

var brokersConfigsGetCmd = &cobra.Command{
        Use:   "get <broker-id>",
        Short: "Show config for broker",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                brokerID := args[0]
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                configs, err := client.GetBrokerConfig(brokerID)
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                headers := []string{"Config Key", "Value"}
                var rows [][]string

                for key, value := range configs {
                        rows = append(rows, []string{key, value})
                }

                formatter.OutputTable(headers, rows)
                return nil
        },
}

var brokersConfigsSetCmd = &cobra.Command{
        Use:   "set <broker-id> <key>=<value>",
        Short: "Update broker config",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
                brokerID := args[0]
                configPair := args[1]
                
                parts := strings.SplitN(configPair, "=", 2)
                if len(parts) != 2 {
                        return fmt.Errorf("config must be in format key=value")
                }
                
                key := parts[0]
                value := parts[1]
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                if err := client.SetBrokerConfig(brokerID, key, value); err != nil {
                        return err
                }

                fmt.Printf("Broker '%s' config '%s' set to '%s'\n", brokerID, key, value)
                return nil
        },
}

func init() {
        brokersCmd.AddCommand(brokersListCmd)
        brokersCmd.AddCommand(brokersDescribeCmd)
        brokersCmd.AddCommand(brokersMetricsCmd)
        brokersCmd.AddCommand(brokersConfigsCmd)

        // Add completion support
        brokersDescribeCmd.ValidArgsFunction = completeBrokerIDs
        brokersConfigsGetCmd.ValidArgsFunction = completeBrokerIDs
        brokersConfigsSetCmd.ValidArgsFunction = completeBrokerIDs

        // Add config subcommands
        brokersConfigsCmd.AddCommand(brokersConfigsListCmd)
        brokersConfigsCmd.AddCommand(brokersConfigsGetCmd)
        brokersConfigsCmd.AddCommand(brokersConfigsSetCmd)
}