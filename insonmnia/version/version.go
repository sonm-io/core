package version

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/mitchellh/go-homedir"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	versionFilePath = "~/.sonm/version.yaml"
	versionURI      = "https://api.github.com/repos/sonm-io/core/git/refs/tags"
)

// Version shows the current SONM platform version.
//
// Components like Node, Worker, CLI, etc. share the same version number.
var Version string

type Observer interface {
	OnError(err error)
	OnDeprecatedVersion(version, latestVersion semver.Version)
	OnBleedingEdgeVersion(version, latestVersion semver.Version)
}

type LogObserver struct {
	log *zap.SugaredLogger
}

func NewLogObserver(log *zap.SugaredLogger) *LogObserver {
	return &LogObserver{
		log: log,
	}
}

func (m *LogObserver) OnError(err error) {
	m.log.Warnw("failed to validate version", zap.Error(err))
}

func (m *LogObserver) OnDeprecatedVersion(version, latestVersion semver.Version) {
	m.log.Warnf("current version %s is OUTDATED, consider update to %s", version.String(), latestVersion.String())
}

func (m *LogObserver) OnBleedingEdgeVersion(version, latestVersion semver.Version) {
	m.log.Warnf("current version %s is UNSTABLE, consider using stable %s version", version.String(), latestVersion.String())
}

// ValidateVersion performs the current component's version validation.
// The idea is to notify whether the current version is either outdated or
// unstable.
func ValidateVersion(ctx context.Context, observer Observer) {
	path, err := homedir.Expand(versionFilePath)
	if err != nil {
		observer.OnError(err)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	manager := NewVersionManager(&fs{path: path})
	if err := manager.ValidateCurrentVersion(ctx); err != nil {
		switch e := err.(type) {
		case *OSError:
			observer.OnError(e)
		case *VersionMismatchError:
			if e.IsDeprecated() {
				observer.OnDeprecatedVersion(e.Version, e.LatestVersion)
			}
			if e.IsBleedingEdge() {
				observer.OnBleedingEdgeVersion(e.Version, e.LatestVersion)
			}
		default:
		}
	}
}

// OSError represents an OS error that is occurred during version checking.
type OSError struct {
	error
}

// Unwrap unwraps the underlying error.
//
// This is useful for further error kind checking, for example using "os.IsPermission(err)".
func (m *OSError) Unwrap() error {
	return m.error
}

// CodecError represents an en/decoding error while deserializing a version
// file content into structured representation.
type CodecError struct {
	error
}

// UpdateError is an error that can occur while updating the latest known
// version.
type UpdateError struct {
	error
}

// VersionMismatchError indicates about version mismatch with the latest one.
type VersionMismatchError struct {
	Version       semver.Version
	LatestVersion semver.Version
}

func (m *VersionMismatchError) IsDeprecated() bool {
	return m.Version.LT(m.LatestVersion)
}

func (m *VersionMismatchError) IsBleedingEdge() bool {
	return m.Version.GT(m.LatestVersion)
}

func (m *VersionMismatchError) Error() string {
	return fmt.Sprintf("version mismatch: %s != %s", m.Version.String(), m.LatestVersion.String())
}

// Option is a type alias for configuring version manager.
type Option func(o *options)

// WithClock allows to assign a specific clock to the version manager.
//
// By default the "time.Now" is used.
func WithClock(clock func() time.Time) Option {
	return func(o *options) {
		o.Clock = clock
	}
}

// WithVersionFetcher allows to specify version fetcher.
//
// By default the GitHub tag fetcher is used.
func WithVersionFetcher(versionFetcher VersionFetcher) Option {
	return func(o *options) {
		o.VersionFetcher = versionFetcher
	}
}

// WithVersion allows to specify an application version.
//
// By default the version obtained through link parameters is used.
func WithVersion(version semver.Version) Option {
	return func(o *options) {
		o.Version = version
	}
}

type options struct {
	Version        semver.Version
	Clock          func() time.Time
	ExpireDuration time.Duration
	VersionFetcher VersionFetcher
}

func newOptions() *options {
	version, err := semver.ParseTolerant(Version)
	if err != nil {
		// This should never happen unless we change the version format
		// in Makefile.
		panic(fmt.Sprintf("failed to parse linked application version: %v", err))
	}

	version.Build = nil
	version.Pre = nil

	return &options{
		Version:        version,
		Clock:          time.Now,
		ExpireDuration: 24 * time.Hour,
		VersionFetcher: newVersionFetcher(versionURI),
	}
}

// File represents a file handle.
//
// Exists primarily for testing purposes.
type File interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close() error
}

// VersionFetcher is a thing that can update the current latest version number
// from external sources.
type VersionFetcher interface {
	Update(ctx context.Context) (semver.Version, error)
}

type gitTag struct {
	Ref string
}

type versionFetcher struct {
	url string
}

func newVersionFetcher(url string) *versionFetcher {
	return &versionFetcher{
		url: url,
	}
}

func (m *versionFetcher) Update(ctx context.Context) (semver.Version, error) {
	response, err := http.Get(m.url)
	if err != nil {
		return semver.Version{}, err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return semver.Version{}, err
	}

	var tags []*gitTag
	if err := json.Unmarshal(data, &tags); err != nil {
		return semver.Version{}, err
	}

	if len(tags) == 0 {
		return semver.Version{}, fmt.Errorf("no versions found")
	}

	version := strings.Replace(tags[len(tags)-1].Ref, "refs/tags/v", "", -1)

	return semver.Parse(version)
}

// FS encapsulates platform dependent filesystem functions.
type FS interface {
	// Open opens the version file for reading.
	Open() (File, error)
	// Create creates the version file with mode 0666.
	Create() (File, error)
}

type fs struct {
	path string
}

func (m *fs) Open() (File, error) {
	return os.Open(m.path)
}

func (m *fs) Create() (File, error) {
	if err := os.MkdirAll(filepath.Dir(m.path), 0750); err != nil {
		return nil, err
	}

	return os.Create(m.path)
}

type versionData struct {
	Version semver.Version
	Updated time.Time
}

type versionRawData struct {
	Version string
	Updated string
}

func (m *versionData) MarshalYAML() (interface{}, error) {
	return &versionRawData{
		Version: m.Version.String(),
		Updated: m.Updated.Format(time.RFC3339),
	}, nil
}

func (m *versionData) UnmarshalYAML(decoder func(interface{}) error) error {
	var content versionRawData
	if err := decoder(&content); err != nil {
		return &CodecError{error: err}
	}

	if len(content.Version) == 0 {
		return &CodecError{error: fmt.Errorf(`missing "version" field`)}
	}

	if len(content.Updated) == 0 {
		return &CodecError{error: fmt.Errorf(`missing "updated" field`)}
	}

	if strings.HasPrefix(content.Version, "v") {
		content.Version = content.Version[1:]
	}

	version, err := semver.Parse(content.Version)
	if err != nil {
		return &CodecError{error: err}
	}

	m.Version = version

	updated, err := time.Parse(time.RFC3339, content.Updated)
	if err != nil {
		return &CodecError{error: err}
	}

	m.Updated = updated

	return nil
}

type VersionManager struct {
	fs             FS
	clock          func() time.Time
	version        semver.Version
	versionFetcher VersionFetcher
	expireDuration time.Duration
}

func NewVersionManager(fs FS, options ...Option) *VersionManager {
	opts := newOptions()
	for _, o := range options {
		o(opts)
	}

	m := &VersionManager{
		fs:             fs,
		clock:          opts.Clock,
		version:        opts.Version,
		versionFetcher: opts.VersionFetcher,
		expireDuration: opts.ExpireDuration,
	}

	return m
}

// ValidateCurrentVersion validates that the current application version is
// the latest, reporting errors otherwise.
//
// Is is possible that this method fails to open the version file for some
// reasons, for example because of permissions. It's the user's responsibility
// either to handle or to ignore the error returned. It's a good idea is to log
// such errors as warnings, but do not terminate the program.
func (m *VersionManager) ValidateCurrentVersion(ctx context.Context) error {
	latestVersion, err := m.latestVersion(ctx)
	if err != nil {
		return err
	}

	if m.version.Compare(latestVersion) != 0 {
		return &VersionMismatchError{
			Version:       m.version,
			LatestVersion: latestVersion,
		}
	}

	return nil
}

func (m *VersionManager) latestVersion(ctx context.Context) (semver.Version, error) {
	versionDataValue, err := m.versionData()
	if err != nil {
		return semver.Version{}, err
	}

	if versionDataValue == nil {
		versionDataValue = &versionData{}
	}

	if m.isExpired(versionDataValue.Updated) {
		version, err := m.versionFetcher.Update(ctx)
		if err != nil {
			return semver.Version{}, &UpdateError{error: err}
		}

		versionDataValue = &versionData{
			Version: version,
			Updated: m.clock(),
		}

		file, err := m.fs.Create()
		if err != nil {
			return semver.Version{}, &OSError{error: err}
		}
		defer file.Close()

		data, err := yaml.Marshal(versionDataValue)
		if err != nil {
			return semver.Version{}, &CodecError{error: err}
		}

		n, err := file.Write(data)
		if err != nil {
			return semver.Version{}, &OSError{error: err}
		}
		if n < len(data) {
			return semver.Version{}, &OSError{error: io.ErrShortWrite}
		}
	}

	return versionDataValue.Version, nil
}

func (m *VersionManager) versionData() (*versionData, error) {
	file, err := m.fs.Open()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, &OSError{error: err}
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, &OSError{error: err}
	}

	var content versionData
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, &CodecError{error: err}
	}

	return &content, nil
}

func (m *VersionManager) isExpired(time time.Time) bool {
	return time.Add(m.expireDuration).Before(m.clock())
}
