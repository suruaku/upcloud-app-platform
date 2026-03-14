package config

import "testing"

func TestValidateAllowsEmptySSHPrivateKeyPath(t *testing.T) {
	t.Parallel()

	cfg := Default()
	cfg.SSH.PrivateKeyPath = ""

	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate config: %v", err)
	}
}
