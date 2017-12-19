package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/distribution/reference"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
)

type Whitelist interface {
	Allowed(name string, digest string) (bool, error)
}

func NewWhitelist(ctx context.Context, config *WhitelistConfig) (Whitelist, error) {
	if !*config.Enabled {
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

func (w *whitelist) Allowed(name string, digest string) (bool, error) {
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

type disabledWhitelist struct {
}

func (w *disabledWhitelist) Allowed(name string, digest string) (bool, error) {
	return true, nil
}
