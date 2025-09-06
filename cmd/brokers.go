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
        "kaf/internal/ai"
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

var (
        metricsAnalyzeFlag  bool
        metricsProviderFlag string
        metricsModelFlag    string
)

var brokersMetricsCmd = &cobra.Command{
        Use:   "metrics <broker-id>",
        Short: "Show broker metrics (disk, CPU, requests)",
        Long:  "Show broker metrics from Prometheus endpoint with optional AI-powered analysis",
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
                        return fmt.Errorf("broker metrics port not configured for current cluster. Use: kaf config update <cluster> --broker-metrics-port <port>")
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

                return fetchAndDisplayMetrics(targetBroker.Host, cluster.BrokerMetricsPort, metricsAnalyzeFlag, metricsProviderFlag, metricsModelFlag)
        },
}

// fetchAndDisplayMetrics retrieves and parses Prometheus metrics from a broker
func fetchAndDisplayMetrics(brokerHost string, metricsPort int, analyze bool, provider, model string) error {
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

        // Perform AI analysis if requested
        if analyze {
                fmt.Println("\n" + strings.Repeat("=", 80))
                fmt.Println("🤖 AI ANALYSIS & RECOMMENDATIONS")
                fmt.Println(strings.Repeat("=", 80))
                
                if err := performAIAnalysis(metrics, provider, model); err != nil {
                        fmt.Printf("⚠️  AI Analysis failed: %v\n", err)
                        fmt.Println("💡 Tip: Make sure your API key is configured for the selected provider")
                }
        }

        return nil
}


// parsePrometheusMetrics parses Prometheus format metrics
func parsePrometheusMetrics(resp *http.Response) []ai.Metric {
        var metrics []ai.Metric
        scanner := bufio.NewScanner(resp.Body)

        for scanner.Scan() {
                line := strings.TrimSpace(scanner.Text())

                // Skip HELP, TYPE, comments and empty lines
                if strings.HasPrefix(line, "#") || line == "" {
                        continue
                }

                metric := parseMetricLine(line)
                if metric != nil {
                        metrics = append(metrics, *metric)
                }
        }

        return metrics
}

// parseMetricLine parses a single Prometheus metric line
func parseMetricLine(line string) *ai.Metric {
        // Find the last space to separate value from metric name + labels
        lastSpaceIndex := strings.LastIndex(line, " ")
        if lastSpaceIndex == -1 {
                return nil
        }

        metricPart := strings.TrimSpace(line[:lastSpaceIndex])
        value := strings.TrimSpace(line[lastSpaceIndex+1:])

        // Parse metric name and labels
        var name, labels string
        
        // Check if there are labels (contains '{' and '}')
        if braceStart := strings.Index(metricPart, "{"); braceStart != -1 {
                name = metricPart[:braceStart]
                
                // Find the closing brace
                braceEnd := strings.LastIndex(metricPart, "}")
                if braceEnd > braceStart {
                        labels = metricPart[braceStart+1:braceEnd]
                }
        } else {
                name = metricPart
                labels = ""
        }

        return &ai.Metric{
                Name:   name,
                Labels: labels,
                Value:  value,
        }
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
                "process_open_fds",
                "process_max_fds",
                "process_virtual_memory_",
        }

        for _, prefix := range kafkaMetricPrefixes {
                if strings.HasPrefix(metricName, prefix) {
                        return true
                }
        }

        return false
}

// performAIAnalysis sends metrics to AI for analysis and displays results
func performAIAnalysis(metrics []ai.Metric, provider, model string) error {
        if provider == "" {
                provider = "openai" // Default provider
        }
        
        // Convert string provider to ai.Provider type
        var aiProvider ai.Provider
        switch strings.ToLower(provider) {
        case "openai":
                aiProvider = ai.OpenAI
        case "claude":
                aiProvider = ai.Claude
        case "grok":
                aiProvider = ai.Grok
        case "gemini":
                aiProvider = ai.Gemini
        default:
                return fmt.Errorf("unsupported AI provider: %s. Available: openai, claude, grok, gemini", provider)
        }
        
        fmt.Printf("🔄 Analyzing metrics with %s...\n\n", strings.ToUpper(provider))
        
        client, err := ai.NewClient(aiProvider)
        if err != nil {
                return err
        }
        
        // Override model if specified
        if model != "" {
                client.Model = model
        }
        
        analysis, err := client.AnalyzeMetrics(metrics)
        if err != nil {
                return err
        }
        
        // Display analysis results
        fmt.Println("📊 SUMMARY:")
        fmt.Printf("   %s\n\n", analysis.Summary)
        
        if len(analysis.Issues) > 0 {
                fmt.Println("⚠️  ISSUES IDENTIFIED:")
                for i, issue := range analysis.Issues {
                        fmt.Printf("   %d. %s\n", i+1, issue)
                }
                fmt.Println()
        }
        
        if len(analysis.RootCauses) > 0 {
                fmt.Println("🔍 ROOT CAUSE ANALYSIS:")
                for i, cause := range analysis.RootCauses {
                        fmt.Printf("   %d. %s\n", i+1, cause)
                }
                fmt.Println()
        }
        
        if len(analysis.Recommendations) > 0 {
                fmt.Println("💡 RECOMMENDATIONS:")
                for i, rec := range analysis.Recommendations {
                        fmt.Printf("   %d. %s\n", i+1, rec)
                }
                fmt.Println()
        }
        
        fmt.Println("ℹ️  Note: AI analysis is based on current metrics snapshot and general best practices.")
        fmt.Println("   Always validate recommendations in your specific environment.")
        
        return nil
}

// displayMetrics formats and displays the parsed metrics
func displayMetrics(metrics []ai.Metric) {
        if len(metrics) == 0 {
                fmt.Println("No Kafka metrics found")
                return
        }

        formatter := getFormatter()
        headers := []string{"Metric", "Labels", "Value"}
        var rows [][]string

        // Sort metrics by name for consistent output
        sort.Slice(metrics, func(i, j int) bool {
                return metrics[i].Name < metrics[j].Name
        })

        for _, metric := range metrics {
                labels := metric.Labels
                if labels == "" {
                        labels = "-"
                }
                
                rows = append(rows, []string{metric.Name, labels, metric.Value})
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

                // Get all keys and sort them alphabetically
                keys := make([]string, 0, len(configs))
                for key := range configs {
                        keys = append(keys, key)
                }
                sort.Strings(keys)

                // Add rows in alphabetical order
                for _, key := range keys {
                        rows = append(rows, []string{key, configs[key]})
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

        // Add AI analysis flags to metrics command
        brokersMetricsCmd.Flags().BoolVarP(&metricsAnalyzeFlag, "analyze", "a", false, "Enable AI-powered analysis and recommendations")
        brokersMetricsCmd.Flags().StringVarP(&metricsProviderFlag, "provider", "p", "openai", "AI provider (openai, claude, grok, gemini)")
        brokersMetricsCmd.Flags().StringVarP(&metricsModelFlag, "model", "m", "", "AI model to use (defaults: gpt-4o, claude-3-sonnet, grok-beta, gemini-pro)")

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