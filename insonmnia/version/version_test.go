package version

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func init() {
	Version = "v0.4.0-f6461c2a"
}

func TestVersionManagerErrFileFailedToOpen(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(nil, os.ErrPermission)

	m := NewVersionManager(fs)
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	osErr, ok := err.(*OSError)
	require.True(t, ok)
	require.True(t, os.IsPermission(osErr.Unwrap()))
}

func TestVersionManagerErrFileRead(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Read(gomock.Any()).Return(0, os.ErrPermission)
	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)

	m := NewVersionManager(fs)
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	osErr, ok := err.(*OSError)
	require.True(t, ok)
	require.True(t, os.IsPermission(osErr.Unwrap()))
}

func TestVersionManagerErrFileInvalidFormat(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	content := `}{`

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Read(gomock.Any()).Do(func(buf []byte) {
		copy(buf, []byte(content))
	}).Return(len(content), io.EOF)
	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)

	m := NewVersionManager(fs)
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	osErr, ok := err.(*CodecError)
	require.True(t, ok)
	require.Error(t, osErr)
}

func TestVersionManagerErrFileMissingVersion(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	content := `{"updated": "2018-11-27 09:55:21+00:00"}`

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Read(gomock.Any()).Do(func(buf []byte) {
		copy(buf, []byte(content))
	}).Return(len(content), io.EOF)
	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)

	m := NewVersionManager(fs)
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	osErr, ok := err.(*CodecError)
	require.True(t, ok)
	require.Error(t, osErr)
}

func TestVersionManagerErrFileMissingUpdated(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	content := `{"version": "v0.4.16"}`

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Read(gomock.Any()).Do(func(buf []byte) {
		copy(buf, []byte(content))
	}).Return(len(content), io.EOF)
	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)

	m := NewVersionManager(fs)
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	osErr, ok := err.(*CodecError)
	require.True(t, ok)
	require.Error(t, osErr)
}

func TestVersionManagerErrFileInvalidVersion(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	content := `{"version": "vvv", "updated": "2018-11-27 09:55:21+00:00"}`

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Read(gomock.Any()).Do(func(buf []byte) {
		copy(buf, []byte(content))
	}).Return(len(content), io.EOF)
	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)

	m := NewVersionManager(fs)
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	osErr, ok := err.(*CodecError)
	require.True(t, ok)
	require.Error(t, osErr)
}

func TestVersionManagerErrFileInvalidUpdated(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	content := `{"version": "v0.4.16", "updated": "2018-11-27"}`

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Read(gomock.Any()).Do(func(buf []byte) {
		copy(buf, []byte(content))
	}).Return(len(content), io.EOF)
	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)

	m := NewVersionManager(fs)
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	osErr, ok := err.(*CodecError)
	require.True(t, ok)
	require.Error(t, osErr)
}

func TestVersionManagerErrUpdateWhenExpired(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	content := `{"version": "v0.4.16", "updated": "2018-11-20T00:00:00+00:00"}`

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Read(gomock.Any()).Do(func(buf []byte) {
		copy(buf, []byte(content))
	}).Return(len(content), io.EOF)

	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)

	versionFetcher := NewMockVersionFetcher(c)
	versionFetcher.EXPECT().Update(context.Background()).Return(semver.Version{}, errors.New("no"))

	clock := func() time.Time {
		now, err := time.Parse(time.RFC3339, "2018-11-22T00:00:00+00:00")
		require.NoError(t, err)
		return now
	}
	m := NewVersionManager(fs, WithClock(clock), WithVersionFetcher(versionFetcher))
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	osErr, ok := err.(*UpdateError)
	require.True(t, ok)
	require.Error(t, osErr)
}

func TestVersionManagerErrUpdateWhenExpiredThenWhenCreate(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	content := `{"version": "v0.4.16", "updated": "2018-11-20T00:00:00+00:00"}`

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Read(gomock.Any()).Do(func(buf []byte) {
		copy(buf, []byte(content))
	}).Return(len(content), io.EOF)

	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)
	fs.EXPECT().Create().Return(nil, os.ErrPermission)

	versionFetcher := NewMockVersionFetcher(c)
	versionFetcher.EXPECT().Update(context.Background()).Return(semver.Version{Minor: 4, Patch: 17}, nil)

	clock := func() time.Time {
		now, err := time.Parse(time.RFC3339, "2018-11-22T00:00:00+00:00")
		require.NoError(t, err)
		return now
	}
	m := NewVersionManager(fs, WithClock(clock), WithVersionFetcher(versionFetcher))
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	osErr, ok := err.(*OSError)
	require.True(t, ok)
	require.Error(t, osErr)
}

func TestVersionManagerErrUpdateWhenExpiredThenWhenSave(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	content := `{"version": "v0.4.16", "updated": "2018-11-20T00:00:00+00:00"}`
	expectedContent := `version: 0.4.17
updated: "2018-11-22T00:00:00Z"
`

	file := NewMockFile(c)
	file.EXPECT().Close().Times(2)
	file.EXPECT().Read(gomock.Any()).Do(func(buf []byte) {
		copy(buf, []byte(content))
	}).Return(len(content), io.EOF)
	file.EXPECT().Write([]byte(expectedContent)).Return(len(expectedContent), os.ErrPermission)

	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)
	fs.EXPECT().Create().Return(file, nil)

	versionFetcher := NewMockVersionFetcher(c)
	versionFetcher.EXPECT().Update(context.Background()).Return(semver.Version{Minor: 4, Patch: 17}, nil)

	clock := func() time.Time {
		now, err := time.Parse(time.RFC3339, "2018-11-22T00:00:00+00:00")
		require.NoError(t, err)
		return now
	}
	m := NewVersionManager(fs, WithClock(clock), WithVersionFetcher(versionFetcher))
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	osErr, ok := err.(*OSError)
	require.True(t, ok)
	require.Error(t, osErr)
}

func TestVersionManagerErrVersionMismatchAfterUpdate(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	content := `{"version": "v0.4.16", "updated": "2018-11-20T00:00:00+00:00"}`

	file := NewMockFile(c)
	file.EXPECT().Close().Times(2)
	file.EXPECT().Read(gomock.Any()).Do(func(buf []byte) {
		copy(buf, []byte(content))
	}).Return(len(content), io.EOF)
	file.EXPECT().Write(gomock.Any()).Return(48, nil)

	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)
	fs.EXPECT().Create().Return(file, nil)

	versionFetcher := NewMockVersionFetcher(c)
	versionFetcher.EXPECT().Update(context.Background()).Return(semver.Version{Minor: 4, Patch: 17}, nil)

	clock := func() time.Time {
		now, err := time.Parse(time.RFC3339, "2018-11-22T00:00:00+00:00")
		require.NoError(t, err)
		return now
	}
	m := NewVersionManager(fs, WithClock(clock), WithVersionFetcher(versionFetcher))
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	versionErr, ok := err.(*VersionMismatchError)
	require.True(t, ok)
	require.Error(t, versionErr)
}

func TestVersionManagerErrVersionMismatchWhenUpdateNotRequired(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	content := `{"version": "v0.4.16", "updated": "2018-11-22T00:00:00+00:00"}`

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Read(gomock.Any()).Do(func(buf []byte) {
		copy(buf, []byte(content))
	}).Return(len(content), io.EOF)

	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(file, nil)

	versionFetcher := NewMockVersionFetcher(c)

	clock := func() time.Time {
		now, err := time.Parse(time.RFC3339, "2018-11-20T12:00:00+00:00")
		require.NoError(t, err)
		return now
	}
	m := NewVersionManager(fs, WithClock(clock), WithVersionFetcher(versionFetcher))
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	versionErr, ok := err.(*VersionMismatchError)
	require.True(t, ok)
	require.Error(t, versionErr)
}

func TestVersionManagerErrVersionMismatchWhenVersionFileNotExists(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Write(gomock.Any()).Return(48, nil)

	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(nil, os.ErrNotExist)
	fs.EXPECT().Create().Return(file, nil)

	versionFetcher := NewMockVersionFetcher(c)
	versionFetcher.EXPECT().Update(context.Background()).Return(semver.Version{Minor: 4, Patch: 17}, nil)

	clock := func() time.Time {
		now, err := time.Parse(time.RFC3339, "2018-11-20T12:00:00+00:00")
		require.NoError(t, err)
		return now
	}
	m := NewVersionManager(fs, WithClock(clock), WithVersionFetcher(versionFetcher))
	err := m.ValidateCurrentVersion(context.Background())

	require.Error(t, err)
	versionErr, ok := err.(*VersionMismatchError)
	require.True(t, ok)
	require.Error(t, versionErr)
}

func TestVersionManager(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	file := NewMockFile(c)
	file.EXPECT().Close().Times(1)
	file.EXPECT().Write(gomock.Any()).Return(48, nil)

	fs := NewMockFS(c)
	fs.EXPECT().Open().Return(nil, os.ErrNotExist)
	fs.EXPECT().Create().Return(file, nil)

	versionFetcher := NewMockVersionFetcher(c)
	versionFetcher.EXPECT().Update(context.Background()).Return(semver.Version{Minor: 4, Patch: 17}, nil)

	clock := func() time.Time {
		now, err := time.Parse(time.RFC3339, "2018-11-20T12:00:00+00:00")
		require.NoError(t, err)
		return now
	}
	m := NewVersionManager(fs, WithVersion(semver.Version{Minor: 4, Patch: 17}), WithClock(clock), WithVersionFetcher(versionFetcher))
	err := m.ValidateCurrentVersion(context.Background())

	require.NoError(t, err)
}
