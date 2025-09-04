package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"kaf/config"
	"kaf/internal/kafka"
)

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage consumer groups",
	Long:  "Commands for managing Kafka consumer groups",
}

var groupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all consumer groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client, err := kafka.NewClient(cfg)
		if err != nil {
			return err
		}

		groups, err := client.ListConsumerGroups()
		if err != nil {
			return err
		}

		formatter := getFormatter()
		headers := []string{"Group ID"}
		var rows [][]string

		for _, group := range groups {
			rows = append(rows, []string{group})
		}

		formatter.OutputTable(headers, rows)
		return nil
	},
}

var groupsDescribeCmd = &cobra.Command{
	Use:   "describe <group>",
	Short: "Show group details (members, offsets)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("describe command not yet implemented")
	},
}

var groupsLagCmd = &cobra.Command{
	Use:   "lag <group>",
	Short: "Show lag metrics per partition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("lag command not yet implemented")
	},
}

var groupsResetCmd = &cobra.Command{
	Use:   "reset <group>",
	Short: "Reset consumer group offsets",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("reset command not yet implemented")
	},
}

var groupsDeleteCmd = &cobra.Command{
	Use:   "delete <group>",
	Short: "Delete a consumer group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("delete command not yet implemented")
	},
}

func init() {
	groupsCmd.AddCommand(groupsListCmd)
	groupsCmd.AddCommand(groupsDescribeCmd)
	groupsCmd.AddCommand(groupsLagCmd)
	groupsCmd.AddCommand(groupsResetCmd)
	groupsCmd.AddCommand(groupsDeleteCmd)

	// Add flags for reset command
	groupsResetCmd.Flags().Bool("to-earliest", false, "Reset to earliest offset")
	groupsResetCmd.Flags().Bool("to-latest", false, "Reset to latest offset")
	groupsResetCmd.Flags().String("to-timestamp", "", "Reset to timestamp")
}