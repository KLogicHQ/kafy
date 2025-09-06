package config

import (
        "fmt"
        "os"
        "path/filepath"

        "gopkg.in/yaml.v3"
)

type Security struct {
        SASL *SASL `yaml:"sasl,omitempty"`
        SSL  bool  `yaml:"ssl,omitempty"`
}

type SASL struct {
        Mechanism string `yaml:"mechanism"`
        Username  string `yaml:"username"`
        Password  string `yaml:"password"`
}

type Cluster struct {
        Bootstrap          string     `yaml:"bootstrap"`
        Zookeeper          string     `yaml:"zookeeper,omitempty"`
        BrokerMetricsPort  int        `yaml:"broker-metrics-port,omitempty"`
        Security           *Security  `yaml:"security,omitempty"`
}

type Config struct {
        CurrentContext string              `yaml:"current-context"`
        Clusters       map[string]*Cluster `yaml:"clusters"`
}

func DefaultConfigPath() string {
        home, err := os.UserHomeDir()
        if err != nil {
                return ".kaf/config.yml"
        }
        return filepath.Join(home, ".kaf", "config.yml")
}

func LoadConfig() (*Config, error) {
        configPath := DefaultConfigPath()
        
        // Create default config if it doesn't exist
        if _, err := os.Stat(configPath); os.IsNotExist(err) {
                return createDefaultConfig(configPath)
        }

        data, err := os.ReadFile(configPath)
        if err != nil {
                return nil, fmt.Errorf("failed to read config file: %w", err)
        }

        var config Config
        if err := yaml.Unmarshal(data, &config); err != nil {
                return nil, fmt.Errorf("failed to parse config file: %w", err)
        }

        if config.Clusters == nil {
                config.Clusters = make(map[string]*Cluster)
        }

        return &config, nil
}

func (c *Config) Save() error {
        configPath := DefaultConfigPath()
        
        // Ensure directory exists
        dir := filepath.Dir(configPath)
        if err := os.MkdirAll(dir, 0755); err != nil {
                return fmt.Errorf("failed to create config directory: %w", err)
        }

        data, err := yaml.Marshal(c)
        if err != nil {
                return fmt.Errorf("failed to marshal config: %w", err)
        }

        if err := os.WriteFile(configPath, data, 0644); err != nil {
                return fmt.Errorf("failed to write config file: %w", err)
        }

        return nil
}

func (c *Config) GetCurrentCluster() (*Cluster, error) {
        if c.CurrentContext == "" {
                return nil, fmt.Errorf("no current context set")
        }

        cluster, exists := c.Clusters[c.CurrentContext]
        if !exists {
                return nil, fmt.Errorf("current context '%s' not found", c.CurrentContext)
        }

        return cluster, nil
}

func createDefaultConfig(configPath string) (*Config, error) {
        config := &Config{
                CurrentContext: "",
                Clusters:       make(map[string]*Cluster),
        }

        // Ensure directory exists
        dir := filepath.Dir(configPath)
        if err := os.MkdirAll(dir, 0755); err != nil {
                return nil, fmt.Errorf("failed to create config directory: %w", err)
        }

        if err := config.Save(); err != nil {
                return nil, err
        }

        return config, nil
}