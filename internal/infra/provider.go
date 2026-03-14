package infra

import (
	"context"
	"time"
)

const UpCloudTokenEnv = "UPCLOUD_TOKEN"

type ProvisionRequest struct {
	Zone         string
	Plan         string
	Template     string
	Hostname     string
	CloudInitRaw []byte
}

type ProvisionResult struct {
	ServerID string
	Hostname string
}

type ServerInfo struct {
	ServerID   string
	Hostname   string
	State      string
	PublicIPv4 string
}

type Provider interface {
	Provision(ctx context.Context, req ProvisionRequest) (ProvisionResult, error)
	Get(ctx context.Context, serverID string) (ServerInfo, error)
	WaitReady(ctx context.Context, serverID string, timeout time.Duration) (ServerInfo, error)
	Destroy(ctx context.Context, serverID string) error
}
