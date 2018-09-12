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
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type Whitelist interface {
	Allowed(ctx context.Context, ref reference.Reference, auth string) (bool, reference.Reference, error)
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

func (w *whitelist) Allowed(ctx context.Context, ref reference.Reference, authority string) (bool, reference.Reference, error) {
	wallet, err := auth.ExtractWalletFromContext(ctx)
	if err != nil {
		log.G(ctx).Warn("could not extract wallet from context", zap.Error(err))
		return false, nil, err
	}
	_, ok := w.superusers[wallet.Hex()]
	if ok {
		return true, ref, nil
	}

	digestedRef, isDigested := ref.(reference.Digested)
	if isDigested {
		if err != nil {
			return false, nil, err
		}
		if _, ok := ref.(reference.Named); !ok {
			return false, nil, errors.New("can not check whitelist for unnamed reference")
		}

		allowed, err := w.digestAllowed(ref.(reference.Named).Name(), (string)(digestedRef.Digest()))

		return allowed, ref, err
	}
	dockerClient, err := dc.NewEnvClient()
	if err != nil {
		return false, nil, err
	}
	defer dockerClient.Close()

	inspection, err := dockerClient.DistributionInspect(ctx, ref.String(), authority)
	if err != nil {
		return false, nil, fmt.Errorf("could not perform DistributionInspect for %s: %v", ref.String(), err)
	}

	ref, err = reference.WithDigest(ref.(reference.Named), inspection.Descriptor.Digest)
	if err != nil {
		// This should never happen
		panic("logical error - can not append digest and parse")
	}

	allowed, err := w.digestAllowed(ref.(reference.Named).Name(), (string)(inspection.Descriptor.Digest))

	return allowed, ref, err
}
