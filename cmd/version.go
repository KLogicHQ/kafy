package cmd

import (
        "fmt"
        "runtime"
        "time"

        "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
        Use:   "version",
        Short: "Show version information",
        Long:  "Display version and build information for kafy CLI",
        RunE: func(cmd *cobra.Command, args []string) error {
                fmt.Printf("kafy version %s\n", version)
                fmt.Printf("Go version: %s\n", runtime.Version())
                fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
                fmt.Printf("Copyright (c) %d KLogic.io. All rights reserved.\n", time.Now().Year())
                fmt.Printf("License: MIT\n")
                return nil
        },
}

func init() {
        rootCmd.AddCommand(versionCmd)
}