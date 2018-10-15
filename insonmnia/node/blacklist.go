package node

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/proto"
)

type blacklistAPI struct {
	remotes *remoteOptions
}

func newBlacklistAPI(opts *remoteOptions) sonm.BlacklistServer {
	return &blacklistAPI{
		remotes: opts,
	}
}

func (m *blacklistAPI) List(ctx context.Context, addr *sonm.EthAddress) (*sonm.BlacklistReply, error) {
	if addr.IsZero() {
		return nil, errors.New("address is empty")
	}

	return m.remotes.dwh.GetBlacklist(ctx, &sonm.BlacklistRequest{UserID: addr})
}

func (m *blacklistAPI) Remove(ctx context.Context, addr *sonm.EthAddress) (*sonm.Empty, error) {
	if err := m.remotes.eth.Blacklist().Remove(ctx, m.remotes.key, addr.Unwrap()); err != nil {
		return nil, fmt.Errorf("failed to remove address from blacklist: %v", err)
	}

	return &sonm.Empty{}, nil
}

func (m *blacklistAPI) Purge(ctx context.Context, req *sonm.Empty) (*sonm.ErrorByStringID, error) {
	myAddr := crypto.PubkeyToAddress(m.remotes.key.PublicKey)
	list, err := m.List(ctx, sonm.NewEthAddress(myAddr))
	if err != nil {
		return nil, err
	}

	concurrency := purgeConcurrency
	if len(list.GetAddresses()) < concurrency {
		concurrency = len(list.GetAddresses())
	}

	status := sonm.NewTSErrorByStringID()
	ch := make(chan *sonm.EthAddress)
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for addr := range ch {
				_, err := m.Remove(ctx, addr)
				status.Append(addr.Unwrap().Hex(), err)
			}

		}()
	}
	for _, rawAddr := range list.GetAddresses() {
		addr, _ := sonm.NewEthAddressFromHex(rawAddr)
		ch <- addr
	}
	close(ch)
	wg.Wait()
	return status.Unwrap(), nil
}
