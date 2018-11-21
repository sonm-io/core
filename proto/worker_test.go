package sonm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestMarshalTaskTag(t *testing.T) {
	tag := &TaskTag{}
	data, err := yaml.Marshal(tag)
	assert.NoError(t, err)
	assert.Equal(t, "null\n", string(data))

	tag.Data = []byte("some")
	data, err = yaml.Marshal(tag)
	assert.NoError(t, err)
	assert.Equal(t, "some\n", string(data))

	tag.Data = []byte{0xff, 0xff}
	data, err = yaml.Marshal(tag)
	assert.NoError(t, err)
	assert.Equal(t, "- 255\n- 255\n", string(data))
}

func TestUnmarshalTaskTag(t *testing.T) {
	data := []byte("[255, 255]")
	reciever := &TaskTag{}
	err := yaml.Unmarshal(data, reciever)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(reciever.Data))
	assert.Equal(t, uint8(255), reciever.Data[0])
	assert.Equal(t, uint8(255), reciever.Data[1])

	strData := "\"some\""
	err = yaml.Unmarshal([]byte(strData), reciever)
	assert.NoError(t, err)
	assert.Equal(t, "some", string(reciever.Data))
}

func TestMarshalUnmarshalTaskTag(t *testing.T) {
	cases := [][]byte{
		[]byte("some"),
		{0xff, 0xff},
		[]byte("c29tZQo="),
		[]byte("\xff"),
		{1, 2, 3},
	}
	for _, cs := range cases {
		initial := &TaskTag{cs}
		receiver := &TaskTag{}
		data, err := yaml.Marshal(initial)
		assert.NoError(t, err)
		err = yaml.Unmarshal(data, receiver)
		assert.NoError(t, err)
		assert.ElementsMatch(t, initial.Data, receiver.Data)
	}
}

func TestMetricsAppend(t *testing.T) {
	m := &WorkerMetricsResponse{}
	x1 := map[string]float64{"x1": 123}
	x2 := map[string]float64{"x2": 123}
	x3 := map[string]float64{"x3": 123}

	m.Append(x1, x2).Append(x3)
	assert.Len(t, m.Metrics, 3)
}
