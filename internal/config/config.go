package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Project   string          `yaml:"project"`
	UpCloud   UpCloudConfig   `yaml:"upcloud"`
	Provision ProvisionConfig `yaml:"provision"`
	SSH       SSHConfig       `yaml:"ssh"`
	Deploy    DeployConfig    `yaml:"deploy"`
}

type UpCloudConfig struct {
	Zone     string `yaml:"zone"`
	Plan     string `yaml:"plan"`
	Template string `yaml:"template"`
}

type ProvisionConfig struct {
	CloudInitPath string `yaml:"cloud_init_path"`
	Hostname      string `yaml:"hostname"`
}

type SSHConfig struct {
	User                  string `yaml:"user"`
	PrivateKeyPath        string `yaml:"private_key_path"`
	ConnectTimeoutSeconds int    `yaml:"connect_timeout_seconds"`
}

type DeployConfig struct {
	ContainerName           string `yaml:"container_name"`
	Image                   string `yaml:"image"`
	Port                    string `yaml:"port"`
	EnvFile                 string `yaml:"env_file"`
	HealthcheckURL          string `yaml:"healthcheck_url"`
	HealthcheckTimeoutSecs  int    `yaml:"healthcheck_timeout_seconds"`
	HealthcheckIntervalSecs int    `yaml:"healthcheck_interval_seconds"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is nil")
	}

	var problems []string
	require := func(name, value string) {
		if strings.TrimSpace(value) == "" {
			problems = append(problems, fmt.Sprintf("%s is required", name))
		}
	}

	require("project", c.Project)
	require("upcloud.zone", c.UpCloud.Zone)
	require("upcloud.plan", c.UpCloud.Plan)
	require("upcloud.template", c.UpCloud.Template)
	require("provision.cloud_init_path", c.Provision.CloudInitPath)
	require("provision.hostname", c.Provision.Hostname)
	require("ssh.user", c.SSH.User)
	require("ssh.private_key_path", c.SSH.PrivateKeyPath)
	require("deploy.container_name", c.Deploy.ContainerName)
	require("deploy.image", c.Deploy.Image)
	require("deploy.port", c.Deploy.Port)
	require("deploy.healthcheck_url", c.Deploy.HealthcheckURL)

	if c.SSH.ConnectTimeoutSeconds <= 0 {
		problems = append(problems, "ssh.connect_timeout_seconds must be > 0")
	}
	if c.Deploy.HealthcheckTimeoutSecs <= 0 {
		problems = append(problems, "deploy.healthcheck_timeout_seconds must be > 0")
	}
	if c.Deploy.HealthcheckIntervalSecs <= 0 {
		problems = append(problems, "deploy.healthcheck_interval_seconds must be > 0")
	}

	if c.Deploy.Port != "" && !strings.Contains(c.Deploy.Port, ":") {
		problems = append(problems, "deploy.port must be in host:container format")
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid config: %s", strings.Join(problems, "; "))
	}

	return nil
}

func Default() Config {
	return Config{
		Project: "my-app",
		UpCloud: UpCloudConfig{
			Zone:     "fi-hel1",
			Plan:     "1xCPU-2GB",
			Template: "Ubuntu Server 24.04 LTS",
		},
		Provision: ProvisionConfig{
			CloudInitPath: "./cloud-init.yaml",
			Hostname:      "my-app-prod",
		},
		SSH: SSHConfig{
			User:                  "ubuntu",
			PrivateKeyPath:        "~/.ssh/id_ed25519",
			ConnectTimeoutSeconds: 120,
		},
		Deploy: DeployConfig{
			ContainerName:           "my-app",
			Image:                   "nginxdemos/hello:latest",
			Port:                    "80:80",
			EnvFile:                 ".env.prod",
			HealthcheckURL:          "http://localhost/",
			HealthcheckTimeoutSecs:  60,
			HealthcheckIntervalSecs: 3,
		},
	}
}

func MarshalYAML(cfg Config) ([]byte, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal config yaml: %w", err)
	}
	return data, nil
}

func EnsureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create parent dir %q: %w", dir, err)
	}
	return nil
}
