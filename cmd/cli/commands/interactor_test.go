package commands

import (
	"fmt"
	"testing"
	"time"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/node"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc/grpclog"
)

var key *ecdsa.PrivateKey

func init() {
	grpclog.SetLogger(logging.NewNullGRPCLogger())

	key, _ = crypto.GenerateKey()
	_, TLSConfig, _ := util.NewHitlessCertRotator(context.Background(), key)
	creds = util.NewTLS(TLSConfig)
}

func initHubItr(t *testing.T) *hubInteractor {
	addr := "127.0.0.1:9999"

	cc, err := util.MakeGrpcClient(context.Background(), addr, creds)
	require.NoError(t, err)

	return &hubInteractor{
		timeout: time.Second,
		hub:     pb.NewHubManagementClient(cc),
	}
}

func initNode(t *testing.T, ctrl *gomock.Controller) *node.Node {
	conf := node.NewMockConfig(ctrl)
	conf.EXPECT().ListenAddress().AnyTimes().Return("127.0.0.1:9999")
	conf.EXPECT().MarketEndpoint().AnyTimes().Return("127.0.0.1:9090")
	conf.EXPECT().LocatorEndpoint().AnyTimes().Return("127.0.0.1:9095")
	conf.EXPECT().HubEndpoint().AnyTimes().Return("")

	n, err := node.New(context.Background(), conf, key)
	require.NoError(t, err)

	return n
}

func TestWrapError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	it := initHubItr(t)
	nod := initNode(t, ctrl)

	go nod.Serve()
	time.Sleep(500 * time.Millisecond)

	typicalError := fmt.Errorf("some non-gRPC error")
	assert.EqualError(t, it.wrapError(typicalError), typicalError.Error())

	_, grpcError := it.hub.Status(context.Background(), &pb.Empty{})
	assert.EqualError(t, it.wrapError(grpcError), errNoHubConnection.Error())
}
