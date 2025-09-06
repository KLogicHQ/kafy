package cmd

import (
        "bufio"
        "fmt"
        "net/http"
        "sort"
        "strconv"
        "strings"
        "time"

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
        Use:   "metrics <broker-id>",
        Short: "Show broker metrics from Prometheus endpoint",
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

                cluster, err := cfg.GetCurrentCluster()
                if err != nil {
                        return err
                }

                if cluster.BrokerMetricsPort == 0 {
                        return fmt.Errorf("broker metrics port not configured for current cluster. Use: kaf config add <cluster> --broker-metrics-port <port>")
                }

                client, err := kafka.NewClient(cfg)
                if err != nil {
                        return err
                }

                brokers, err := client.ListBrokers()
                if err != nil {
                        return err
                }

                // Find the broker
                var targetBroker *kafka.BrokerInfo
                for _, broker := range brokers {
                        if broker.ID == int32(brokerID) {
                                targetBroker = &broker
                                break
                        }
                }

                if targetBroker == nil {
                        return fmt.Errorf("broker %d not found", brokerID)
                }

                return fetchAndDisplayMetrics(targetBroker.Host, cluster.BrokerMetricsPort)
        },
}

// fetchAndDisplayMetrics retrieves and parses Prometheus metrics from a broker
func fetchAndDisplayMetrics(brokerHost string, metricsPort int) error {
        url := fmt.Sprintf("http://%s:%d/metrics", brokerHost, metricsPort)
        
        client := &http.Client{
                Timeout: 10 * time.Second,
        }
        
        resp, err := client.Get(url)
        if err != nil {
                return fmt.Errorf("failed to fetch metrics from %s: %w", url, err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != http.StatusOK {
                return fmt.Errorf("metrics endpoint returned status %d", resp.StatusCode)
        }
        
        metrics := parsePrometheusMetrics(resp)
        displayMetrics(metrics)
        
        return nil
}

// parsePrometheusMetrics parses Prometheus format metrics
func parsePrometheusMetrics(resp *http.Response) map[string]string {
        metrics := make(map[string]string)
        scanner := bufio.NewScanner(resp.Body)
        
        for scanner.Scan() {
                line := strings.TrimSpace(scanner.Text())
                
                // Skip comments and empty lines
                if strings.HasPrefix(line, "#") || line == "" {
                        continue
                }
                
                // Parse metric line: metric_name{labels} value
                parts := strings.Fields(line)
                if len(parts) >= 2 {
                        metricName := parts[0]
                        value := parts[1]
                        
                        // Only include key Kafka metrics
                        if isKafkaMetric(metricName) {
                                metrics[metricName] = value
                        }
                }
        }
        
        return metrics
}

// isKafkaMetric checks if a metric is relevant for Kafka monitoring
func isKafkaMetric(metricName string) bool {
        kafkaMetricPrefixes := []string{
                "kafka_server_",
                "kafka_controller_",
                "kafka_coordinator_",
                "kafka_log_",
                "kafka_network_",
                "jvm_memory_",
                "jvm_gc_",
                "process_cpu_",
                "process_resident_memory_",
        }
        
        for _, prefix := range kafkaMetricPrefixes {
                if strings.HasPrefix(metricName, prefix) {
                        return true
                }
        }
        
        return false
}

// displayMetrics formats and displays the parsed metrics
func displayMetrics(metrics map[string]string) {
        if len(metrics) == 0 {
                fmt.Println("No Kafka metrics found")
                return
        }
        
        formatter := getFormatter()
        headers := []string{"Metric", "Value"}
        var rows [][]string
        
        // Sort metrics for consistent output
        var sortedKeys []string
        for key := range metrics {
                sortedKeys = append(sortedKeys, key)
        }
        sort.Strings(sortedKeys)
        
        for _, key := range sortedKeys {
                // Clean up metric name for display
                displayName := strings.ReplaceAll(key, "_", " ")
                displayName = strings.Title(displayName)
                
                rows = append(rows, []string{displayName, metrics[key]})
        }
        
        formatter.OutputTable(headers, rows)
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
        brokersMetricsCmd.ValidArgsFunction = completeBrokerIDs
        brokersConfigsGetCmd.ValidArgsFunction = completeBrokerIDs
        brokersConfigsSetCmd.ValidArgsFunction = completeBrokerIDs

        // Add config subcommands
        brokersConfigsCmd.AddCommand(brokersConfigsListCmd)
        brokersConfigsCmd.AddCommand(brokersConfigsGetCmd)
        brokersConfigsCmd.AddCommand(brokersConfigsSetCmd)
}