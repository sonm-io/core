package xdocker

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReferenceMarshalUnmarshal(t *testing.T) {
	refStr := "httpd:latest"
	ref, err := NewReference(refStr)
	require.NoError(t, err)
	data, err := json.Marshal(&ref)
	require.NoError(t, err)
	require.Equal(t, "\"docker.io/library/httpd:latest\"", string(data))

	ref2 := &Reference{}
	err = json.Unmarshal(data, ref2)
	require.NoError(t, err)
	require.Equal(t, "docker.io/library/httpd:latest", ref2.String())
}
