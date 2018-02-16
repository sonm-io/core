package hub

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func clusterClient(ctrl *gomock.Controller) Cluster {
	c := NewMockCluster(ctrl)
	c.EXPECT().IsLeader().AnyTimes().Return(true)
	c.EXPECT().Members().MinTimes(1).AnyTimes().Return([]NewMemberEvent{}, nil)
	return c
}

func TestAnnouncerOK(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	c := clusterClient(ctrl)
	_, key := makeTestKey()

	lc := sonm.NewMockLocatorClient(ctrl)
	lc.EXPECT().Announce(gomock.Any(), gomock.Any()).MinTimes(2).Return(&sonm.Empty{}, nil)

	ann := newLocatorAnnouncer(key, lc, time.Second, c)
	// announce once, look at error
	err := ann.Once(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "", ann.ErrorMsg())

	// start announcer, wait a bit for sending
	go func() { ann.Start(ctx) }()
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, "", ann.ErrorMsg())
}

func TestAnnouncerHasError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	c := clusterClient(ctrl)
	_, key := makeTestKey()
	lc := sonm.NewMockLocatorClient(ctrl)
	lc.EXPECT().Announce(gomock.Any(), gomock.Any()).MinTimes(2).
		Return(nil, errors.New("test: cannot announce"))

	ann := newLocatorAnnouncer(key, lc, time.Second, c)

	err := ann.Once(ctx)
	assert.EqualError(t, err, "test: cannot announce")
	assert.Equal(t, "test: cannot announce", ann.ErrorMsg())

	// start announcer, wait a bit for sending
	go func() { ann.Start(ctx) }()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, "test: cannot announce", ann.ErrorMsg())
}
