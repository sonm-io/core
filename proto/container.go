package sonm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
)

func (m *Registry) Auth() string {
	if m == nil {
		return ""
	}

	data, err := json.Marshal(m.authConfig())
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(data)
}

func (m *Registry) authConfig() types.AuthConfig {
	return types.AuthConfig{
		Username:      m.GetUsername(),
		Password:      m.GetPassword(),
		ServerAddress: m.GetServerAddress(),
	}
}

func (m *Container) Validate() error {
	if m.GetImage() == "" {
		return fmt.Errorf("container image name is required")
	}

	return nil
}
