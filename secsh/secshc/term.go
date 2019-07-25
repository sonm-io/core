package secshc

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sync/errgroup"
)

type RemotePTY struct {
	cfg        *RPTYConfig
	privateKey *ecdsa.PrivateKey
}

// NewRemotePTY constructs a new remote PTY (pseudo terminal) over a secured
// connection.
func NewRemotePTY(cfg *RPTYConfig) (*RemotePTY, error) {
	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load ETH keys: %v", err)
	}

	m := &RemotePTY{
		cfg:        cfg,
		privateKey: key,
	}

	return m, nil
}

// Run runs the PTY by capturing the standard input and executing commands
// on remote host over the secured connection.
func (m *RemotePTY) Run(ctx context.Context, addr auth.Addr) error {
	certRotator, tlsConfig, err := util.NewHitlessCertRotator(ctx, m.privateKey)
	if err != nil {
		return err
	}

	defer certRotator.Close()

	nppDialerOptions := []npp.Option{
		npp.WithProtocol(Protocol),
		npp.WithRendezvous(m.cfg.NPP.Rendezvous, xgrpc.NewTransportCredentials(tlsConfig)),
		npp.WithLogger(zap.NewNop()),
	}
	nppDialer, err := npp.NewDialer(nppDialerOptions...)
	if err != nil {
		return err
	}

	conn, err := nppDialer.DialContext(ctx, addr)
	if err != nil {
		return err
	}

	ethAddr, err := addr.ETH()
	if err != nil {
		return err
	}

	cc, err := xgrpc.NewClient(ctx, "-", auth.NewWalletAuthenticator(xgrpc.NewTransportCredentials(tlsConfig), ethAddr), xgrpc.WithConn(conn))
	if err != nil {
		return err
	}

	remotePTY := sonm.NewRemotePTYClient(cc)

	banner, err := remotePTY.Banner(ctx, &sonm.RemotePTYBannerRequest{})
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", banner.Banner)

	stdin := newStdinReader()

	tty := terminal.NewTerminal(stdin, "$: ")

	return m.withRaw(func(ttyState *terminal.State) error {
		for {
			line, err := tty.ReadLine()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}

			if line == "exit" {
				return nil
			}

			fn := func() {
				if err := executeCmd(ctx, remotePTY, line, stdin); err != nil {
					fmt.Printf("%v\n\r", err)
				}
			}

			if err := m.withRestored(ttyState, fn); err != nil {
				return err
			}
		}
	})
}

func (m *RemotePTY) withRaw(fn func(ttyState *terminal.State) error) error {
	ttyState, err := terminal.MakeRaw(0)
	if err != nil {
		return err
	}

	defer func() {
		if err := terminal.Restore(0, ttyState); err != nil {
			fmt.Printf("ERROR: %v\n\r", err)
		}
	}()

	return fn(ttyState)
}

func (m *RemotePTY) withRestored(ttyState *terminal.State, fn func()) error {
	if err := terminal.Restore(0, ttyState); err != nil {
		return err
	}

	fn()

	if _, err := terminal.MakeRaw(0); err != nil {
		return err
	}

	return nil
}

func executeCmd(ctx context.Context, remotePTY sonm.RemotePTYClient, line string, stdin *stdinPipe) error {
	if len(line) == 0 {
		return nil
	}

	stream, err := remotePTY.Exec(ctx, &sonm.RemotePTYExecRequest{
		Args: strings.Fields(line),
		Envp: nil,
	})

	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	readContext, readContextCancel := context.WithCancel(ctx)

	wg.Go(func() error {
		defer readContextCancel()

		for {
			chunk, err := stream.Recv()

			// Yeah, dat shit when we receive BOTH error and non-nil data.
			if chunk != nil {
				fmt.Printf("%s", string(chunk.Out))
			}

			switch err {
			case nil:
				break
			case io.EOF:
				return nil
			default:
				return err
			}

			if chunk.Done {
				return nil
			}
		}
	})
	wg.Go(func() error {
		buf := []byte{0}
		for {
			n, err := stdin.ReadContext(readContext, buf)
			if err != nil {
				return nil
			}

			if n == 0 {
				continue
			}

			fmt.Printf("%v\n", buf[0])
		}
	})

	return wg.Wait()
}

type stdinPipe struct {
	mu sync.Mutex
	ch chan byteOrError
}

func newStdinReader() *stdinPipe {
	ch := make(chan byteOrError, 1)
	go func() {
		defer close(ch)

		b := []byte{0}
		for {
			n, err := os.Stdin.Read(b)
			if err != nil {
				ch <- byteOrError{err: err}
				return
			}

			if n > 0 {
				ch <- byteOrError{b: b[0]}
			}
		}
	}()

	return &stdinPipe{
		ch: ch,
	}
}

func (m *stdinPipe) Read(p []byte) (n int, err error) {
	return m.ReadContext(context.Background(), p)
}

func (m *stdinPipe) Write(p []byte) (n int, err error) {
	return os.Stdin.Write(p)
}

func (m *stdinPipe) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case v := <-m.ch:
		if v.err != nil {
			return 0, v.err
		}

		p[0] = v.b
		return 1, nil
	}
}

type byteOrError struct {
	b   byte
	err error
}
