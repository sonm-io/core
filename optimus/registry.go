package optimus

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type certificate struct {
	rotator util.HitlessCertRotator
	cfg     *tls.Config
	cred    credentials.TransportCredentials
}

// Registry keeps all used components in itself.
type Registry struct {
	mu           sync.Mutex
	certificates map[common.Address]*certificate
	connections  []*grpc.ClientConn
}

func newRegistry() *Registry {
	return &Registry{
		certificates: map[common.Address]*certificate{},
	}
}

func (m *Registry) NewWorkerManagement(ctx context.Context, addr auth.Addr, privateKey *ecdsa.PrivateKey) (sonm.WorkerManagementClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, err := m.newClient(ctx, addr, privateKey)
	if err != nil {
		return nil, err
	}

	return sonm.NewWorkerManagementClient(conn), nil
}

func (m *Registry) NewDWH(ctx context.Context, addr auth.Addr, privateKey *ecdsa.PrivateKey) (sonm.DWHClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, err := m.newClient(ctx, addr, privateKey)
	if err != nil {
		return nil, err
	}

	return sonm.NewDWHClient(conn), nil
}

func (m *Registry) newClient(ctx context.Context, addr auth.Addr, privateKey *ecdsa.PrivateKey, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	cred, err := m.credentials(ctx, privateKey)
	if err != nil {
		return nil, err
	}

	conn, err := xgrpc.NewClient(ctx, addr.String(), cred, opts...)
	if err != nil {
		return nil, err
	}

	m.connections = append(m.connections, conn)

	return conn, nil
}

func (m *Registry) credentials(ctx context.Context, privateKey *ecdsa.PrivateKey) (credentials.TransportCredentials, error) {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	if certificate, ok := m.certificates[addr]; ok {
		return certificate.cred, nil
	}

	cert, TLSConfig, err := util.NewHitlessCertRotator(ctx, privateKey)
	if err != nil {
		return nil, err
	}

	cred := util.NewTLS(TLSConfig)
	m.certificates[addr] = &certificate{
		rotator: cert,
		cfg:     TLSConfig,
		cred:    cred,
	}

	return cred, nil
}

func (m *Registry) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, certificate := range m.certificates {
		certificate.rotator.Close()
	}

	for _, conn := range m.connections {
		conn.Close()
	}
}
