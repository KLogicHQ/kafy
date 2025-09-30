package main

import (
        "os"

        "kafy/cmd"
)

func main() {
        if err := cmd.Execute(); err != nil {
                // Error is already printed by cobra with usage/help
                // Just exit with error code
                os.Exit(1)
        }
}