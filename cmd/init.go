package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local project configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: initialize config at", cfgFile)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
