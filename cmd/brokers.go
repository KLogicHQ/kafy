package cmd

import (
        "fmt"
        "strconv"

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

func init() {
        brokersCmd.AddCommand(brokersListCmd)
        brokersCmd.AddCommand(brokersDescribeCmd)
        brokersCmd.AddCommand(brokersMetricsCmd)
}