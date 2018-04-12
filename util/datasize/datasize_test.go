package datasize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBitRate_UnmarshalText(t *testing.T) {
	text := []byte("300 Mbit/s")
	rate := BitRate{}
	err := rate.UnmarshalText(text)
	require.NoError(t, err)
	marshalled, err := rate.MarshalText()
	require.NoError(t, err)
	require.Equal(t, text, marshalled)

	text = []byte("3445 Mbit/s")
	err = rate.UnmarshalText(text)
	require.NoError(t, err)
	hr := rate.HumanReadableDec()
	require.Equal(t, hr, "3.445 Gbit/s")

	text = []byte("3.445 Gbit/s")
	err = rate.UnmarshalText(text)
	require.NoError(t, err)
	hr = rate.HumanReadableDec()
	require.Equal(t, hr, string(text))

	text = []byte("3445 Mbit")
	err = rate.UnmarshalText(text)
	require.Error(t, err)

	text = []byte("1 MB/s")
	err = rate.UnmarshalText(text)
	require.NoError(t, err)
	require.Equal(t, rate.Bits(), uint64(8*1000*1000))

	text = []byte("1Mb/s")
	err = rate.UnmarshalText(text)
	require.NoError(t, err)
	require.Equal(t, rate.Bits(), uint64(1e6))

	text = []byte("100")
	err = rate.UnmarshalText(text)
	require.NoError(t, err)
	require.Equal(t, rate.Bits(), uint64(100))
}

func TestByteSize_UnmarshalText(t *testing.T) {
	text := []byte("300 MB")
	rate := ByteSize{}
	err := rate.UnmarshalText(text)
	require.NoError(t, err)
	marshalled, err := rate.MarshalText()
	require.NoError(t, err)
	require.Equal(t, text, marshalled)

	text = []byte("3445 MB")
	err = rate.UnmarshalText(text)
	require.NoError(t, err)
	hr := rate.HumanReadableDec()
	require.Equal(t, hr, "3.445 GB")

	text = []byte("3445 Mbit/s")
	err = rate.UnmarshalText(text)
	require.Error(t, err)

	text = []byte("3445 Mb")
	err = rate.UnmarshalText(text)
	require.Error(t, err)

	text = []byte("1MB")
	err = rate.UnmarshalText(text)
	require.NoError(t, err)
	require.Equal(t, rate.Bytes(), uint64(1e6))

	text = []byte("100")
	err = rate.UnmarshalText(text)
	require.NoError(t, err)
	require.Equal(t, rate.Bytes(), uint64(100))

}
