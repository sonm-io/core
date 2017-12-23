package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/distribution/reference"
	dc "github.com/docker/docker/client"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type Whitelist interface {
	Allowed(ctx context.Context, registry string, image string, auth string) (bool, reference.Named, error)
}

func NewWhitelist(ctx context.Context, config *WhitelistConfig) Whitelist {
	if config.Enabled != nil && !*config.Enabled {
		return &disabledWhitelist{}
	}

	wl := whitelist{}

	for _, su := range config.PrivilegedAddresses {
		wl.superusers[su] = struct{}{}
	}

	go wl.updateRoutine(ctx, config.Url, config.RefreshPeriod)

	return &wl
}

type WhitelistRecord struct {
	AllowedHashes []string `json:"allowed_hashes"`
}

type whitelist struct {
	superusers map[string]struct{}
	Records    map[string]WhitelistRecord
	RecordsMu  sync.RWMutex
}

func (w *whitelist) updateRoutine(ctx context.Context, url string, updatePeriod uint) error {
	ticker := util.NewImmediateTicker(time.Duration(time.Second * time.Duration(updatePeriod)))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := w.load(ctx, url)
			if err != nil {
				log.G(ctx).Error("could not load whitelist", zap.Error(err))
			}
		}
	}
}

func (w *whitelist) load(ctx context.Context, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.G(ctx).Info("fetched whitelist")
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download whitelist - got %s", resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)
	r := make(map[string]WhitelistRecord)
	err = decoder.Decode(&r)
	if err != nil {
		return errors.Wrap(err, "could not decode whitelist data")
	}

	w.RecordsMu.Lock()
	w.Records = r
	w.RecordsMu.Unlock()

	return nil
}

func (w *whitelist) digestAllowed(name string, digest string) (bool, error) {
	ref, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return false, err
	}

	w.RecordsMu.RLock()
	defer w.RecordsMu.RUnlock()
	record, ok := w.Records[ref.Name()]
	if !ok {
		return false, nil
	}

	for _, allowedDigest := range record.AllowedHashes {
		if allowedDigest == digest {
			return true, nil
		}
	}

	return false, nil
}

func (w *whitelist) Allowed(ctx context.Context, registry string, image string, auth string) (bool, reference.Named, error) {
	fullName := filepath.Join(registry, image)
	ref, err := reference.ParseNormalizedNamed(fullName)
	if err != nil {
		return false, nil, err
	}

	wallet, err := extractWalletFromContext(ctx)
	if err != nil {
		log.G(ctx).Warn("could not extract wallet from context", zap.Error(err))
	}
	_, ok := w.superusers[wallet]
	if ok {
		return true, ref, nil
	}

	digestedRef, isDigested := ref.(reference.Digested)
	if isDigested {
		if err != nil {
			return false, nil, err
		}

		allowed, err := w.digestAllowed(ref.Name(), (string)(digestedRef.Digest()))

		return allowed, ref, err
	}
	dockerClient, err := dc.NewEnvClient()
	if err != nil {
		return false, nil, err
	}
	defer dockerClient.Close()

	inspection, err := dockerClient.DistributionInspect(ctx, image, auth)
	if err != nil {
		return false, nil, errors.Wrap(err, "could not perform DistributionInspect")
	}

	ref, err = reference.ParseNormalizedNamed(ref.String() + "@" + (string)(inspection.Descriptor.Digest))
	if err != nil {
		// This should never happen
		panic("logical error - can not append digest and parse")
	}

	allowed, err := w.digestAllowed(ref.Name(), (string)(inspection.Descriptor.Digest))

	return allowed, ref, err
}

type disabledWhitelist struct {
}

func (w *disabledWhitelist) Allowed(ctx context.Context, registry string, image string, auth string) (bool, reference.Named, error) {
	fullName := filepath.Join(registry, image)
	ref, err := reference.ParseNormalizedNamed(fullName)
	if err != nil {
		return false, nil, err
	}

	return true, ref, nil
}
