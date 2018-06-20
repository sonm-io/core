// +build linux

package storagequota

import (
	"context"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

func TestBTRFSQuota(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)
	dclient, err := client.NewEnvClient()
	require.NoError(err)
	defer dclient.Close()

	info, err := dclient.Info(ctx)
	require.NoError(err)

	if !QuotationSupported(info) {
		t.Skipf("quota is not supported with %s", info.Driver)
	}
}
