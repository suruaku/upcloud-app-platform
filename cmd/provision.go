package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var provisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision secure Docker host on UpCloud",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: provision server using", cfgFile)
	},
}

func init() {
	rootCmd.AddCommand(provisionCmd)
}
