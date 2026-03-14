package ssh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRunnerAutoDetectsPrivateKeyWhenPathEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatalf("create ssh dir: %v", err)
	}
	ed25519Path := filepath.Join(sshDir, "id_ed25519")
	if err := os.WriteFile(ed25519Path, []byte("test-key"), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	runner, err := NewRunner(Config{User: "ubuntu", PrivateKeyPath: ""})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	if runner.privateKeyPath != ed25519Path {
		t.Fatalf("private key path = %q, want %q", runner.privateKeyPath, ed25519Path)
	}
}

func TestNewRunnerAutoDetectsFallbackKeyOrder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatalf("create ssh dir: %v", err)
	}
	ecdsaPath := filepath.Join(sshDir, "id_ecdsa")
	if err := os.WriteFile(ecdsaPath, []byte("test-key"), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	runner, err := NewRunner(Config{User: "ubuntu", PrivateKeyPath: ""})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	if runner.privateKeyPath != ecdsaPath {
		t.Fatalf("private key path = %q, want %q", runner.privateKeyPath, ecdsaPath)
	}
}

func TestNewRunnerEmptyPathFailsWhenNoDefaultKeysFound(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatalf("create ssh dir: %v", err)
	}

	_, err := NewRunner(Config{User: "ubuntu", PrivateKeyPath: ""})
	if err == nil {
		t.Fatal("expected error")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "ssh.private_key_path") {
		t.Fatalf("error %q does not mention ssh.private_key_path", errMsg)
	}
	if !strings.Contains(errMsg, "id_ed25519") || !strings.Contains(errMsg, "id_ecdsa") || !strings.Contains(errMsg, "id_rsa") {
		t.Fatalf("error %q does not list attempted default keys", errMsg)
	}
}

func TestNewRunnerExplicitInvalidPathFailsFast(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatalf("create ssh dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("test-key"), 0o600); err != nil {
		t.Fatalf("write fallback key: %v", err)
	}

	_, err := NewRunner(Config{User: "ubuntu", PrivateKeyPath: "./does-not-exist"})
	if err == nil {
		t.Fatal("expected error")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "does-not-exist") {
		t.Fatalf("error %q does not include invalid configured path", errMsg)
	}
}
