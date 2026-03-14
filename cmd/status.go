package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show infrastructure and app status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: show status using", cfgFile)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
