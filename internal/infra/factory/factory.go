package factory

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ikox01/upcloud-box/internal/infra"
	upcloudapi "github.com/ikox01/upcloud-box/internal/infra/upcloudapi"
)

func NewDefaultProvider() (infra.Provider, error) {
	token := strings.TrimSpace(os.Getenv(infra.UpCloudTokenEnv))
	if token == "" {
		return nil, fmt.Errorf("missing required environment variable %s", infra.UpCloudTokenEnv)
	}

	provider, err := upcloudapi.NewProvider(token, 60*time.Second)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
