package upcloudapi

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v8/upcloud/service"
	"github.com/ikox01/upcloud-box/internal/infra"
)

var uuidPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

type Provider struct {
	svc *service.Service
}

func NewProvider(token string, timeout time.Duration) (*Provider, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("upcloud token is required")
	}

	cl := client.New("", "", client.WithBearerAuth(token), client.WithTimeout(timeout))
	return &Provider{svc: service.New(cl)}, nil
}

func (p *Provider) Provision(ctx context.Context, req infra.ProvisionRequest) (infra.ProvisionResult, error) {
	templateUUID, err := p.resolveTemplateUUID(ctx, req.Zone, req.Template)
	if err != nil {
		return infra.ProvisionResult{}, err
	}

	created, err := p.svc.CreateServer(ctx, &request.CreateServerRequest{
		Zone:             req.Zone,
		Title:            req.Hostname,
		Hostname:         req.Hostname,
		Plan:             req.Plan,
		PasswordDelivery: request.PasswordDeliveryNone,
		Metadata:         upcloud.True,
		NICModel:         upcloud.NICModelVirtio,
		UserData:         string(req.CloudInitRaw),
		Networking: &request.CreateServerNetworking{
			Interfaces: []request.CreateServerInterface{
				{
					Type: upcloud.IPAddressAccessPublic,
					IPAddresses: []request.CreateServerIPAddress{
						{Family: upcloud.IPAddressFamilyIPv4},
					},
				},
			},
		},
		StorageDevices: []request.CreateServerStorageDevice{
			{
				Action:  request.CreateServerStorageDeviceActionClone,
				Storage: templateUUID,
				Title:   req.Hostname + "-disk-1",
			},
		},
	})
	if err != nil {
		return infra.ProvisionResult{}, fmt.Errorf("create upcloud server: %w", err)
	}

	return infra.ProvisionResult{ServerID: created.UUID, Hostname: created.Hostname}, nil
}

func (p *Provider) Get(ctx context.Context, serverID string) (infra.ServerInfo, error) {
	details, err := p.svc.GetServerDetails(ctx, &request.GetServerDetailsRequest{UUID: serverID})
	if err != nil {
		return infra.ServerInfo{}, fmt.Errorf("get server details %q: %w", serverID, err)
	}

	return toServerInfo(details), nil
}

func (p *Provider) WaitReady(ctx context.Context, serverID string, timeout time.Duration) (infra.ServerInfo, error) {
	waitCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	details, err := p.svc.WaitForServerState(waitCtx, &request.WaitForServerStateRequest{
		UUID:         serverID,
		DesiredState: upcloud.ServerStateStarted,
	})
	if err != nil {
		return infra.ServerInfo{}, fmt.Errorf("wait for server %q to become started: %w", serverID, err)
	}

	return toServerInfo(details), nil
}

func (p *Provider) Destroy(ctx context.Context, serverID string) error {
	if err := p.svc.DeleteServer(ctx, &request.DeleteServerRequest{UUID: serverID}); err != nil {
		return fmt.Errorf("delete server %q: %w", serverID, err)
	}
	return nil
}

func (p *Provider) resolveTemplateUUID(ctx context.Context, zone string, template string) (string, error) {
	template = strings.TrimSpace(template)
	if template == "" {
		return "", fmt.Errorf("template is required")
	}
	if uuidPattern.MatchString(template) {
		return template, nil
	}

	storages, err := p.svc.GetStorages(ctx, &request.GetStoragesRequest{
		Access: upcloud.StorageAccessPublic,
		Type:   upcloud.StorageTypeTemplate,
	})
	if err != nil {
		return "", fmt.Errorf("list public template storages: %w", err)
	}

	for _, s := range storages.Storages {
		if s.Zone == zone && s.Title == template && s.State == upcloud.StorageStateOnline {
			return s.UUID, nil
		}
	}

	needle := strings.ToLower(template)
	for _, s := range storages.Storages {
		if s.Zone == zone && strings.Contains(strings.ToLower(s.Title), needle) && s.State == upcloud.StorageStateOnline {
			return s.UUID, nil
		}
	}

	return "", fmt.Errorf("no storage template match found in zone %q for template %q", zone, template)
}

func toServerInfo(details *upcloud.ServerDetails) infra.ServerInfo {
	info := infra.ServerInfo{
		ServerID: details.UUID,
		Hostname: details.Hostname,
		State:    details.State,
	}

	for _, ip := range details.IPAddresses {
		if ip.Access == upcloud.IPAddressAccessPublic && ip.Family == upcloud.IPAddressFamilyIPv4 {
			info.PublicIPv4 = ip.Address
			break
		}
	}

	return info
}
