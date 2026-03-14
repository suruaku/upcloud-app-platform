package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy single container workload",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: deploy container using", cfgFile)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
