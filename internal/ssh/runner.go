package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	User           string
	PrivateKeyPath string
	ConfigDir      string
	ConnectTimeout time.Duration
	RetryInterval  time.Duration
}

type Runner struct {
	user           string
	privateKeyPath string
	connectTimeout time.Duration
	retryInterval  time.Duration
}

type Result struct {
	Stdout string
	Stderr string
}

func NewRunner(cfg Config) (*Runner, error) {
	user := strings.TrimSpace(cfg.User)
	if user == "" {
		return nil, fmt.Errorf("ssh user is required")
	}

	var keyPath string
	privateKeyPath := strings.TrimSpace(cfg.PrivateKeyPath)
	if privateKeyPath != "" {
		resolvedKeyPath, err := resolvePath(privateKeyPath, cfg.ConfigDir)
		if err != nil {
			return nil, fmt.Errorf("resolve ssh private key path: %w", err)
		}
		if _, err := os.Stat(resolvedKeyPath); err != nil {
			return nil, fmt.Errorf("ssh private key path %q: %w", resolvedKeyPath, err)
		}
		keyPath = resolvedKeyPath
	} else {
		detectedKeyPath, attempted, err := detectPrivateKeyPath()
		if err != nil {
			return nil, err
		}
		if detectedKeyPath == "" {
			return nil, fmt.Errorf("ssh private key path is empty and no default SSH key was found (tried: %s); set ssh.private_key_path", strings.Join(attempted, ", "))
		}
		keyPath = detectedKeyPath
	}

	connectTimeout := cfg.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 30 * time.Second
	}

	retryInterval := cfg.RetryInterval
	if retryInterval <= 0 {
		retryInterval = 3 * time.Second
	}

	return &Runner{
		user:           user,
		privateKeyPath: keyPath,
		connectTimeout: connectTimeout,
		retryInterval:  retryInterval,
	}, nil
}

func detectPrivateKeyPath() (string, []string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, fmt.Errorf("resolve home dir for ssh key detection: %w", err)
	}

	candidates := []string{"id_ed25519", "id_ecdsa", "id_rsa"}
	attempted := make([]string, 0, len(candidates))
	for _, name := range candidates {
		candidate := filepath.Join(home, ".ssh", name)
		attempted = append(attempted, candidate)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, attempted, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", attempted, fmt.Errorf("check ssh private key path %q: %w", candidate, err)
		}
	}

	return "", attempted, nil
}

func (r *Runner) RunWithRetry(ctx context.Context, host, remoteCommand string, timeout time.Duration) (Result, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return Result{}, fmt.Errorf("ssh host is required")
	}

	if strings.TrimSpace(remoteCommand) == "" {
		return Result{}, fmt.Errorf("remote command is required")
	}

	if timeout <= 0 {
		timeout = r.connectTimeout
	}

	checkCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		checkCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	var lastErr error
	for {
		result, err := r.runOnce(checkCtx, host, remoteCommand)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if errors.Is(checkCtx.Err(), context.DeadlineExceeded) {
			return Result{}, fmt.Errorf("ssh command did not succeed within %s: %w", timeout, lastErr)
		}
		if checkCtx.Err() != nil {
			return Result{}, checkCtx.Err()
		}

		timer := time.NewTimer(r.retryInterval)
		select {
		case <-checkCtx.Done():
			timer.Stop()
			if errors.Is(checkCtx.Err(), context.DeadlineExceeded) {
				return Result{}, fmt.Errorf("ssh command did not succeed within %s: %w", timeout, lastErr)
			}
			return Result{}, checkCtx.Err()
		case <-timer.C:
		}
	}
}

func (r *Runner) Run(ctx context.Context, host, remoteCommand string) (Result, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return Result{}, fmt.Errorf("ssh host is required")
	}

	if strings.TrimSpace(remoteCommand) == "" {
		return Result{}, fmt.Errorf("remote command is required")
	}

	return r.runOnce(ctx, host, remoteCommand)
}

func (r *Runner) UploadFile(ctx context.Context, host, localPath, remotePath string) error {
	host = strings.TrimSpace(host)
	if host == "" {
		return fmt.Errorf("ssh host is required")
	}

	localPath = strings.TrimSpace(localPath)
	if localPath == "" {
		return fmt.Errorf("local path is required")
	}

	remotePath = strings.TrimSpace(remotePath)
	if remotePath == "" {
		return fmt.Errorf("remote path is required")
	}

	if _, err := os.Stat(localPath); err != nil {
		return fmt.Errorf("local file %q: %w", localPath, err)
	}

	connectSeconds := int(r.connectTimeout / time.Second)
	if connectSeconds < 1 {
		connectSeconds = 1
	}

	args := []string{
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", fmt.Sprintf("ConnectTimeout=%d", connectSeconds),
		"-i", r.privateKeyPath,
		localPath,
		fmt.Sprintf("%s@%s:%s", r.user, host, remotePath),
	}

	cmd := exec.CommandContext(ctx, "scp", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText == "" {
			errText = strings.TrimSpace(err.Error())
		}
		return fmt.Errorf("upload file %q to %s:%s: %s", localPath, host, remotePath, errText)
	}

	return nil
}

func (r *Runner) runOnce(ctx context.Context, host, remoteCommand string) (Result, error) {
	connectSeconds := int(r.connectTimeout / time.Second)
	if connectSeconds < 1 {
		connectSeconds = 1
	}

	args := []string{
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", fmt.Sprintf("ConnectTimeout=%d", connectSeconds),
		"-i", r.privateKeyPath,
		fmt.Sprintf("%s@%s", r.user, host),
		remoteCommand,
	}

	cmd := exec.CommandContext(ctx, "ssh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText == "" {
			errText = strings.TrimSpace(err.Error())
		}
		return Result{Stdout: stdout.String(), Stderr: stderr.String()}, fmt.Errorf("run ssh command %q on %s: %s", remoteCommand, host, errText)
	}

	return Result{Stdout: stdout.String(), Stderr: stderr.String()}, nil
}

func resolvePath(path, baseDir string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	path = os.ExpandEnv(path)
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve current directory: %w", err)
		}
		baseDir = cwd
	}

	return filepath.Clean(filepath.Join(baseDir, path)), nil
}
