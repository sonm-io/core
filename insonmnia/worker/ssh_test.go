package worker

import (
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHKeyParse(t *testing.T) {
	raw := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh+u6UN26+nIc42aRhnuDeralPivXZDi3ETSugsNlOfMww5YdqSJc9otSGooPRbXhOVguoEZfBvLNNd4xTkYtaCsWmFGbq3JXCjtH22V3VeqDc1zd3iJGtQU2BInC0HHvR4M5U4ayN4Ur3bEwgBViv7J+2lABmOArVwOlxacI/m2FtmUPrXKLh98eZgvAxd7DLwTjL8DKLJVqk2hqPRbqvX+CVHVZ4EeS63k0ji2mHDDlZrCsm2n6CnOau4sIND4Xiibdtt6dHnXKXxyC1SLQlH1W+6fxdiQSWXK4/Q4ryA0L/t89CoSp+/uRy4xnP3z5ntI7vE+I3Y1kFeTpOy1v9")

	pkey := PublicKey{}
	err := pkey.UnmarshalText(raw)
	require.NoError(t, err)

	b, err := pkey.MarshalText()
	require.NoError(t, err)

	parsed, _, _, _, err := ssh.ParseAuthorizedKey(b)
	require.NoError(t, err)
	assert.True(t, ssh.KeysEqual(pkey, parsed))
}
