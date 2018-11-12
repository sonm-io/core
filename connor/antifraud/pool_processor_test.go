package antifraud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/oleiade/lane.v1"
)

func TestPoolProcessorQueue_Decide(t *testing.T) {
	w := &commonPoolProcessor{
		hashrateQueue: &lane.Queue{Deque: lane.NewCappedDeque(60)},
	}

	// empty Q should return true - not enough data to analyze
	assert.True(t, w.nonZeroHashrate())

	// last five non-zero items, hashrate is non-zero
	for i := 0; i <= 5; i++ {
		w.updateHashRateQueue(float64(i))
	}
	assert.True(t, w.nonZeroHashrate())

	// last five zero items - should decide that there is no
	// hashrate
	for i := 0; i < 5; i++ {
		w.updateHashRateQueue(0)
	}
	assert.False(t, w.nonZeroHashrate())
}
