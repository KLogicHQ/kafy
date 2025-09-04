package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var offsetsCmd = &cobra.Command{
	Use:   "offsets",
	Short: "Manage topic offsets",
	Long:  "Commands for managing and inspecting topic offsets",
}

var offsetsShowCmd = &cobra.Command{
	Use:   "show <topic>",
	Short: "Show offsets per partition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("show command not yet implemented")
	},
}

var offsetsResetCmd = &cobra.Command{
	Use:   "reset <topic>",
	Short: "Reset topic offsets",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("reset command not yet implemented")
	},
}

func init() {
	offsetsCmd.AddCommand(offsetsShowCmd)
	offsetsCmd.AddCommand(offsetsResetCmd)

	// Add flags for reset command
	offsetsResetCmd.Flags().Bool("to-earliest", false, "Reset to earliest offset")
	offsetsResetCmd.Flags().Bool("to-latest", false, "Reset to latest offset")
	offsetsResetCmd.Flags().String("to-timestamp", "", "Reset to timestamp")
}