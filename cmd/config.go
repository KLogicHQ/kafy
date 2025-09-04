package cmd

import (
        "fmt"
        "os"

        "github.com/spf13/cobra"
        "gopkg.in/yaml.v3"
        "kaf/config"
)

var configCmd = &cobra.Command{
        Use:   "config",
        Short: "Manage Kafka cluster configurations",
        Long:  "Commands for managing Kafka cluster configurations and contexts",
}

var configListCmd = &cobra.Command{
        Use:   "list",
        Short: "List all configured clusters",
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                
                if len(cfg.Clusters) == 0 {
                        fmt.Println("No clusters configured")
                        return nil
                }

                headers := []string{"Name", "Bootstrap", "Current"}
                var rows [][]string
                
                for name, cluster := range cfg.Clusters {
                        current := ""
                        if name == cfg.CurrentContext {
                                current = "*"
                        }
                        rows = append(rows, []string{name, cluster.Bootstrap, current})
                }

                formatter.OutputTable(headers, rows)
                return nil
        },
}

var configCurrentCmd = &cobra.Command{
        Use:   "current",
        Short: "Show the current active cluster",
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                if cfg.CurrentContext == "" {
                        fmt.Println("No current context set")
                        return nil
                }

                cluster, exists := cfg.Clusters[cfg.CurrentContext]
                if !exists {
                        return fmt.Errorf("current context '%s' not found", cfg.CurrentContext)
                }

                formatter := getFormatter()
                result := map[string]interface{}{
                        "name":      cfg.CurrentContext,
                        "bootstrap": cluster.Bootstrap,
                }
                
                if cluster.Zookeeper != "" {
                        result["zookeeper"] = cluster.Zookeeper
                }

                return formatter.Output(result)
        },
}

var configUseCmd = &cobra.Command{
        Use:   "use <cluster-name>",
        Short: "Switch to a different cluster",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                clusterName := args[0]
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                if _, exists := cfg.Clusters[clusterName]; !exists {
                        return fmt.Errorf("cluster '%s' not found", clusterName)
                }

                cfg.CurrentContext = clusterName
                if err := cfg.Save(); err != nil {
                        return err
                }

                fmt.Printf("Switched to cluster '%s'\n", clusterName)
                return nil
        },
}

var configAddCmd = &cobra.Command{
        Use:   "add <cluster-name>",
        Short: "Add a new cluster",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                clusterName := args[0]
                bootstrap, _ := cmd.Flags().GetString("bootstrap")
                zookeeper, _ := cmd.Flags().GetString("zookeeper")
                
                if bootstrap == "" {
                        return fmt.Errorf("--bootstrap flag is required")
                }

                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                if cfg.Clusters == nil {
                        cfg.Clusters = make(map[string]*config.Cluster)
                }

                cfg.Clusters[clusterName] = &config.Cluster{
                        Bootstrap: bootstrap,
                        Zookeeper: zookeeper,
                }

                // Set as current if it's the first cluster
                if cfg.CurrentContext == "" {
                        cfg.CurrentContext = clusterName
                }

                if err := cfg.Save(); err != nil {
                        return err
                }

                fmt.Printf("Added cluster '%s'\n", clusterName)
                return nil
        },
}

var configDeleteCmd = &cobra.Command{
        Use:   "delete <cluster-name>",
        Short: "Remove a cluster",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                clusterName := args[0]
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                if _, exists := cfg.Clusters[clusterName]; !exists {
                        return fmt.Errorf("cluster '%s' not found", clusterName)
                }

                delete(cfg.Clusters, clusterName)

                // Reset current context if it was deleted
                if cfg.CurrentContext == clusterName {
                        cfg.CurrentContext = ""
                        // Set to another cluster if available
                        for name := range cfg.Clusters {
                                cfg.CurrentContext = name
                                break
                        }
                }

                if err := cfg.Save(); err != nil {
                        return err
                }

                fmt.Printf("Deleted cluster '%s'\n", clusterName)
                return nil
        },
}

var configRenameCmd = &cobra.Command{
        Use:   "rename <old-name> <new-name>",
        Short: "Rename a cluster context",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
                oldName := args[0]
                newName := args[1]
                
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                cluster, exists := cfg.Clusters[oldName]
                if !exists {
                        return fmt.Errorf("cluster '%s' not found", oldName)
                }

                if _, exists := cfg.Clusters[newName]; exists {
                        return fmt.Errorf("cluster '%s' already exists", newName)
                }

                cfg.Clusters[newName] = cluster
                delete(cfg.Clusters, oldName)

                if cfg.CurrentContext == oldName {
                        cfg.CurrentContext = newName
                }

                if err := cfg.Save(); err != nil {
                        return err
                }

                fmt.Printf("Renamed cluster '%s' to '%s'\n", oldName, newName)
                return nil
        },
}

var configExportCmd = &cobra.Command{
        Use:   "export",
        Short: "Export config to YAML/JSON",
        RunE: func(cmd *cobra.Command, args []string) error {
                cfg, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                formatter := getFormatter()
                return formatter.Output(cfg)
        },
}

var configImportCmd = &cobra.Command{
        Use:   "import <file>",
        Short: "Import config from YAML/JSON",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                filePath := args[0]
                
                data, err := os.ReadFile(filePath)
                if err != nil {
                        return fmt.Errorf("failed to read file: %w", err)
                }

                var importedConfig config.Config
                if err := yaml.Unmarshal(data, &importedConfig); err != nil {
                        return fmt.Errorf("failed to parse config file: %w", err)
                }

                // Load existing config
                existingConfig, err := config.LoadConfig()
                if err != nil {
                        return err
                }

                // Merge configs (imported clusters will overwrite existing ones)
                for name, cluster := range importedConfig.Clusters {
                        existingConfig.Clusters[name] = cluster
                }

                // Update current context if it was set in imported config
                if importedConfig.CurrentContext != "" {
                        existingConfig.CurrentContext = importedConfig.CurrentContext
                }

                if err := existingConfig.Save(); err != nil {
                        return err
                }

                fmt.Printf("Successfully imported configuration from %s\n", filePath)
                return nil
        },
}

func init() {
        configCmd.AddCommand(configListCmd)
        configCmd.AddCommand(configCurrentCmd)
        configCmd.AddCommand(configUseCmd)
        configCmd.AddCommand(configAddCmd)
        configCmd.AddCommand(configDeleteCmd)
        configCmd.AddCommand(configRenameCmd)
        configCmd.AddCommand(configExportCmd)
        configCmd.AddCommand(configImportCmd)

        // Add flags
        configAddCmd.Flags().String("bootstrap", "", "Bootstrap servers (required)")
        configAddCmd.Flags().String("zookeeper", "", "Zookeeper connection string")
}