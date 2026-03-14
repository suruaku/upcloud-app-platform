package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrBootstrapConfigForUpCreatesConfigAndCloudInit(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)

	configPath := filepath.Join(workDir, "upcloud-box.yaml")

	cfg, bootstrap, err := loadOrBootstrapConfigForUp(configPath)
	if err != nil {
		t.Fatalf("load or bootstrap config: %v", err)
	}
	if cfg == nil {
		t.Fatal("config is nil")
	}
	if !bootstrap.ConfigCreated {
		t.Fatal("expected config to be created")
	}
	if !bootstrap.CloudInitCreated {
		t.Fatal("expected cloud-init to be created")
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "cloud-init.yaml")); err != nil {
		t.Fatalf("stat cloud-init: %v", err)
	}
}

func TestLoadOrBootstrapConfigForUpKeepsExistingCloudInit(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)

	cloudInitPath := filepath.Join(workDir, "cloud-init.yaml")
	originalCloudInit := []byte("#cloud-config\nusers: []\n")
	if err := os.WriteFile(cloudInitPath, originalCloudInit, 0o600); err != nil {
		t.Fatalf("write cloud-init: %v", err)
	}

	configPath := filepath.Join(workDir, "upcloud-box.yaml")

	_, bootstrap, err := loadOrBootstrapConfigForUp(configPath)
	if err != nil {
		t.Fatalf("load or bootstrap config: %v", err)
	}
	if !bootstrap.ConfigCreated {
		t.Fatal("expected config to be created")
	}
	if bootstrap.CloudInitCreated {
		t.Fatal("expected existing cloud-init to be kept")
	}

	after, err := os.ReadFile(cloudInitPath)
	if err != nil {
		t.Fatalf("read cloud-init: %v", err)
	}
	if string(after) != string(originalCloudInit) {
		t.Fatalf("cloud-init content changed unexpectedly\nwant:\n%s\n\ngot:\n%s", string(originalCloudInit), string(after))
	}
}

func TestLoadOrBootstrapConfigForUpUsesExistingConfig(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)

	configPath := filepath.Join(workDir, "upcloud-box.yaml")
	if err := writeConfig(configPath, "./cloud-init.yaml", false); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, bootstrap, err := loadOrBootstrapConfigForUp(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if bootstrap.ConfigCreated {
		t.Fatal("did not expect config bootstrap")
	}
	if bootstrap.CloudInitCreated {
		t.Fatal("did not expect cloud-init bootstrap")
	}
}
