package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/docker/distribution/reference"
	dc "github.com/docker/docker/client"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
)

type Whitelist interface {
	Allowed(ctx context.Context, registry string, image string, auth string) (bool, reference.Named, error)
}

func NewWhitelist(ctx context.Context, config *WhitelistConfig) (Whitelist, error) {
	if config.Enabled != nil && !*config.Enabled {
		return &disabledWhitelist{}, nil
	}
	resp, err := http.Get(config.Url)
	if err != nil {
		return nil, err
	}
	log.G(ctx).Info("fetched whitelist")
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download whitelist - got %s", resp.Status)
	}
	decoder := json.NewDecoder(resp.Body)
	wl := whitelist{}
	err = decoder.Decode(&wl)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode whitelist data")
	}
	return &wl, nil
}

type WhitelistRecord struct {
	AllowedHashes []string
}

type whitelist struct {
	Records map[string]WhitelistRecord
}

func (w *whitelist) digestAllowed(name string, digest string) (bool, error) {
	ref, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return false, err
	}
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
