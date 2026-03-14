package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ikox01/upcloud-box/internal/infra/factory"
	"github.com/ikox01/upcloud-box/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show infrastructure and app status",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := state.Load(state.DefaultPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf("No state file found at %s\n", state.DefaultPath)
				return nil
			}
			return err
		}

		fmt.Printf("State file: %s\n", state.DefaultPath)
		fmt.Printf("server_uuid: %s\n", renderOrDash(s.ServerUUID))
		fmt.Printf("public_ip: %s\n", renderOrDash(s.PublicIP))
		fmt.Printf("last_successful_image: %s\n", renderOrDash(s.LastSuccessfulImage))
		fmt.Printf("last_deployed_at: %s\n", renderOrDash(s.LastDeployedAt))

		if strings.TrimSpace(s.ServerUUID) == "" {
			fmt.Println("Remote infra: none tracked")
			return nil
		}

		provider, err := factory.NewDefaultProvider()
		if err != nil {
			fmt.Printf("Remote infra: skipped (%v)\n", err)
			return nil
		}

		serverInfo, err := provider.Get(context.Background(), s.ServerUUID)
		if err != nil {
			if isLikelyNotFound(err) {
				fmt.Printf("Remote infra: server %s not found\n", s.ServerUUID)
				return nil
			}
			return err
		}

		fmt.Printf("Remote infra: %s (%s)\n", serverInfo.ServerID, serverInfo.Hostname)
		fmt.Printf("remote_state: %s\n", renderOrDash(serverInfo.State))
		fmt.Printf("remote_public_ipv4: %s\n", renderOrDash(serverInfo.PublicIPv4))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func renderOrDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

func isLikelyNotFound(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") || strings.Contains(msg, "status code 404") || strings.Contains(msg, " 404")
}
