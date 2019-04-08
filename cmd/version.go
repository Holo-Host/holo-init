package cmd

import (
  "fmt"

  "github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
  Use:   "version",
  Short: "Holo cli version number",
  Long:  `Every software program has a version, and this is the Holo cli version`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Holo cli v0.0.1")
  },
}
