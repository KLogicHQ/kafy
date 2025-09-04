package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"kaf/config"
	"kaf/internal/kafka"
)

var topicsCmd = &cobra.Command{
	Use:   "topics",
	Short: "Manage Kafka topics",
	Long:  "Commands for managing Kafka topics",
}

var topicsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all topics",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client, err := kafka.NewClient(cfg)
		if err != nil {
			return err
		}

		topics, err := client.ListTopics()
		if err != nil {
			return err
		}

		formatter := getFormatter()
		headers := []string{"Name", "Partitions", "Replicas"}
		var rows [][]string

		for _, topic := range topics {
			rows = append(rows, []string{
				topic.Name,
				strconv.Itoa(topic.Partitions),
				strconv.Itoa(topic.Replicas),
			})
		}

		formatter.OutputTable(headers, rows)
		return nil
	},
}

var topicsDescribeCmd = &cobra.Command{
	Use:   "describe <topic>",
	Short: "Show details of a topic",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		topicName := args[0]
		
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client, err := kafka.NewClient(cfg)
		if err != nil {
			return err
		}

		topic, err := client.DescribeTopic(topicName)
		if err != nil {
			return err
		}

		formatter := getFormatter()
		return formatter.Output(topic)
	},
}

var topicsCreateCmd = &cobra.Command{
	Use:   "create <topic>",
	Short: "Create a topic",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		topicName := args[0]
		partitions, _ := cmd.Flags().GetInt("partitions")
		replication, _ := cmd.Flags().GetInt("replication")
		
		if partitions <= 0 {
			return fmt.Errorf("partitions must be greater than 0")
		}
		if replication <= 0 {
			return fmt.Errorf("replication factor must be greater than 0")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client, err := kafka.NewClient(cfg)
		if err != nil {
			return err
		}

		if err := client.CreateTopic(topicName, partitions, replication); err != nil {
			return err
		}

		fmt.Printf("Topic '%s' created successfully\n", topicName)
		return nil
	},
}

var topicsDeleteCmd = &cobra.Command{
	Use:   "delete <topic>",
	Short: "Delete a topic (with safety confirmation)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		topicName := args[0]
		force, _ := cmd.Flags().GetBool("force")
		
		if !force {
			fmt.Printf("Are you sure you want to delete topic '%s'? (y/N): ", topicName)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client, err := kafka.NewClient(cfg)
		if err != nil {
			return err
		}

		if err := client.DeleteTopic(topicName); err != nil {
			return err
		}

		fmt.Printf("Topic '%s' deleted successfully\n", topicName)
		return nil
	},
}

var topicsAlterCmd = &cobra.Command{
	Use:   "alter <topic>",
	Short: "Alter partitions or configs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("alter command not yet implemented")
	},
}

func init() {
	topicsCmd.AddCommand(topicsListCmd)
	topicsCmd.AddCommand(topicsDescribeCmd)
	topicsCmd.AddCommand(topicsCreateCmd)
	topicsCmd.AddCommand(topicsDeleteCmd)
	topicsCmd.AddCommand(topicsAlterCmd)

	// Add flags
	topicsCreateCmd.Flags().Int("partitions", 1, "Number of partitions")
	topicsCreateCmd.Flags().Int("replication", 1, "Replication factor")
	topicsDeleteCmd.Flags().Bool("force", false, "Skip confirmation")
}