package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy provisioned resources",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: destroy resources using", cfgFile)
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
