package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/docker/distribution/reference"
	dc "github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/opencontainers/go-digest"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xdocker"
	"go.uber.org/zap"
)

type Whitelist interface {
	Allowed(ctx context.Context, ref xdocker.Reference, auth string) (bool, xdocker.Reference, error)
}

func NewWhitelist(ctx context.Context, config *WhitelistConfig) Whitelist {
	wl := whitelist{
		superusers: make(map[string]struct{}),
	}

	for _, su := range config.PrivilegedAddresses {
		parsed := common.HexToAddress(su)
		wl.superusers[parsed.Hex()] = struct{}{}
	}

	go wl.updateRoutine(ctx, config.Url, config.RefreshPeriod)

	return &wl
}

type WhitelistRecord struct {
	AllowedHashes []digest.Digest `json:"allowed_hashes"`
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

	return w.fillFromJsonReader(ctx, resp.Body)
}

func (w *whitelist) fillFromJsonReader(ctx context.Context, jsonReader io.Reader) error {
	decoder := json.NewDecoder(jsonReader)
	r := make(map[string]WhitelistRecord)
	err := decoder.Decode(&r)
	if err != nil {
		return fmt.Errorf("could not decode whitelist data: %v", err)
	}

	w.RecordsMu.Lock()
	w.Records = r
	w.RecordsMu.Unlock()

	return nil
}

func (w *whitelist) digestAllowed(name string, digest digest.Digest) (bool, error) {
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

func (w *whitelist) Allowed(ctx context.Context, ref xdocker.Reference, authority string) (bool, xdocker.Reference, error) {
	wallet, err := auth.ExtractWalletFromContext(ctx)
	if err != nil {
		log.G(ctx).Warn("could not extract wallet from context", zap.Error(err))
		return false, ref, err
	}
	_, ok := w.superusers[wallet.Hex()]
	if ok {
		return true, ref, nil
	}

	if !ref.HasName() {
		return false, ref, errors.New("can not check whitelist for unnamed reference")
	}

	if ref.HasDigest() {
		allowed, err := w.digestAllowed(ref.Name(), ref.Digest())
		return allowed, ref, err
	}

	dockerClient, err := dc.NewEnvClient()
	if err != nil {
		return false, ref, err
	}
	defer dockerClient.Close()

	inspection, err := dockerClient.DistributionInspect(ctx, ref.String(), authority)
	if err != nil {
		return false, ref, fmt.Errorf("could not perform DistributionInspect for %s: %v", ref.String(), err)
	}

	ref, err = ref.WithDigest(inspection.Descriptor.Digest)
	if err != nil {
		return false, ref, fmt.Errorf("could not add digest to reference %s: %v", ref.String(), err)
	}

	allowed, err := w.digestAllowed(ref.Name(), inspection.Descriptor.Digest)

	return allowed, ref, err
}
