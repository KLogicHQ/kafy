package cmd

import (
        "fmt"
        "runtime"

        "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
        Use:   "version",
        Short: "Show version information",
        Long:  "Display version and build information for kaf CLI",
        RunE: func(cmd *cobra.Command, args []string) error {
                fmt.Printf("kaf version %s\n", version)
                fmt.Printf("Go version: %s\n", runtime.Version())
                fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
                fmt.Printf("Copyright (c) 2024 KLogic.io. All rights reserved.\n")
                fmt.Printf("License: MIT\n")
                return nil
        },
}

func init() {
        rootCmd.AddCommand(versionCmd)
}