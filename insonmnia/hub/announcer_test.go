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

func TestAnnouncerOK(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	_, key := makeTestKey()

	lc := sonm.NewMockLocatorClient(ctrl)
	lc.EXPECT().Announce(gomock.Any(), gomock.Any()).MinTimes(2).Return(&sonm.Empty{}, nil)

	cfg := Config{Endpoint: ":5050"}
	ann, err := newLocatorAnnouncer(key, lc, time.Second, &cfg)
	assert.NoError(t, err)
	// announce once, look at error
	err = ann.Once(ctx)
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

	_, key := makeTestKey()
	lc := sonm.NewMockLocatorClient(ctrl)
	lc.EXPECT().Announce(gomock.Any(), gomock.Any()).MinTimes(2).
		Return(nil, errors.New("test: cannot announce"))

	cfg := Config{Endpoint: ":5050"}
	ann, err := newLocatorAnnouncer(key, lc, time.Second, &cfg)
	assert.NoError(t, err)

	err = ann.Once(ctx)
	assert.EqualError(t, err, "test: cannot announce")
	assert.Equal(t, "test: cannot announce", ann.ErrorMsg())

	// start announcer, wait a bit for sending
	go func() { ann.Start(ctx) }()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, "test: cannot announce", ann.ErrorMsg())
}

func TestGetEndpoints(t *testing.T) {
	assert := assert.New(t)
	var fixtures = []struct {
		Endpoint    string
		ExpectError bool
	}{
		{Endpoint: ":10001", ExpectError: false},
		{Endpoint: "0.0.0.0:10001", ExpectError: false},
		{Endpoint: "aaaa:50000", ExpectError: true},
	}

	for _, fixture := range fixtures {
		clientEndpoints, err := getEndpoints(&Config{Endpoint: fixture.Endpoint})
		if fixture.ExpectError {
			assert.NotNil(err)
		} else {
			assert.NoError(err)
			assert.NotEmpty(clientEndpoints)
		}
	}
}
