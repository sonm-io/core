package hub

import (
	"context"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

func makeStore(ctx context.Context, cfg *HubConfig) (store.Store, error) {
	consul.Register()
	boltdb.Register()
	log.G(ctx).Info("creating store", zap.Any("store", cfg.Store))

	endpoints := []string{cfg.Store.Endpoint}

	backend := store.Backend(cfg.Store.Type)

	config := store.Config{}
	config.Bucket = cfg.Store.Bucket
	return libkv.NewStore(backend, endpoints, &config)
}
