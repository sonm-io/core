package worker

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

var (
	_    = setupTestResponder()
	key  = getTestKey()
	addr = ethcrypto.PubkeyToAddress(key.PublicKey)
)

func getTestKey() *ecdsa.PrivateKey {
	k, _ := ethcrypto.GenerateKey()
	return k
}

func setupTestResponder() *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	l, _ := net.Listen("tcp", "127.0.0.1:4242")
	ts.Listener = l
	ts.Start()

	return ts
}

func TestTransformEnvVars(t *testing.T) {
	vars := map[string]string{
		"key1": "value1",
		"KEY2": "VALUE2",
		"keY":  "12345",
		"key4": "",
	}

	description := Description{Container: sonm.Container{Env: vars}}

	assert.Contains(t, description.FormatEnv(), "key1=value1")
	assert.Contains(t, description.FormatEnv(), "KEY2=VALUE2")
	assert.Contains(t, description.FormatEnv(), "keY=12345")
	assert.Contains(t, description.FormatEnv(), "key4=")
}

func TestCollectTasksStatuses(t *testing.T) {
	m := Worker{
		options: &options{
			ctx: context.Background(),
		},
		containers: map[string]*ContainerInfo{
			"aaa1": {status: sonm.TaskStatusReply_UNKNOWN},
			"aaa2": {status: sonm.TaskStatusReply_UNKNOWN},
			"bbb2": {status: sonm.TaskStatusReply_SPOOLING},
			"ccc2": {status: sonm.TaskStatusReply_SPAWNING},
			"ddd1": {status: sonm.TaskStatusReply_RUNNING},
			"ddd2": {status: sonm.TaskStatusReply_RUNNING},
			"ddd3": {status: sonm.TaskStatusReply_RUNNING},
			"eee1": {status: sonm.TaskStatusReply_FINISHED},
			"fff2": {status: sonm.TaskStatusReply_BROKEN},
		},
	}

	result1 := m.CollectTasksStatuses()
	assert.Equal(t, len(m.containers), len(result1))

	result2 := m.CollectTasksStatuses(sonm.TaskStatusReply_UNKNOWN)
	assert.Equal(t, 2, len(result2))

	result3 := m.CollectTasksStatuses(sonm.TaskStatusReply_RUNNING)
	assert.Equal(t, 3, len(result3))

	result4 := m.CollectTasksStatuses(sonm.TaskStatusReply_RUNNING, sonm.TaskStatusReply_SPOOLING, sonm.TaskStatusReply_BROKEN)
	assert.Equal(t, 5, len(result4))
}
