package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ikox01/upcloud-box/internal/config"
	"github.com/ikox01/upcloud-box/internal/state"

	"github.com/spf13/cobra"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local project configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := writeConfig(cfgFile, initForce); err != nil {
			return err
		}

		if err := writeState(state.DefaultPath, initForce); err != nil {
			return err
		}

		fmt.Printf("Created %s\n", cfgFile)
		fmt.Printf("Created %s\n", state.DefaultPath)
		fmt.Println("Next: edit your config values, then run upcloud-box provision")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing files")
	rootCmd.AddCommand(initCmd)
}

func writeConfig(path string, force bool) error {
	if err := config.EnsureParentDir(path); err != nil {
		return err
	}

	if err := ensureWritable(path, force); err != nil {
		return err
	}

	defaultCfg := config.Default()
	data, err := config.MarshalYAML(defaultCfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}

	return nil
}

func writeState(path string, force bool) error {
	if err := ensureWritable(path, force); err != nil {
		return err
	}

	if err := state.Save(path, state.New()); err != nil {
		return err
	}

	return nil
}

func ensureWritable(path string, force bool) error {
	_, err := os.Stat(path)
	if err == nil && !force {
		return fmt.Errorf("file %q already exists (use --force to overwrite)", path)
	}
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("check file %q: %w", path, err)
}
