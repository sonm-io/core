package worker

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oschwald/geoip2-golang"
	"github.com/sonm-io/core/util/multierror"
)

const (
	DefaultDatabasePath = "/usr/local/share/sonm/geoip/geobase.mmdb"
	DefaultDatabaseURL  = "https://s3.eu-west-2.amazonaws.com/sonm.geoip/geobase.mmdb"
	DefaultDatabaseSHA1 = "2ef2075cc6e4567ab4efca3fde906f4611f18d49"
)

type GeoIPServiceConfig struct {
	Path string `yaml:"path"`
	URL  string `yaml:"url"`
	SHA1 string `yaml:"sha1"`
}

func (m *GeoIPServiceConfig) Normalize() error {
	if len(m.Path) == 0 {
		m.Path = DefaultDatabasePath
	}

	if len(m.URL) == 0 {
		m.URL = DefaultDatabaseURL
	}

	if len(m.SHA1) == 0 {
		m.SHA1 = DefaultDatabaseSHA1
	}

	if len(m.SHA1) != 2*sha1.Size {
		return fmt.Errorf("geo IP database SHA1 checksum HEX string must have %d bytes", 2*sha1.Size)
	}

	return nil
}

type GeoIPServiceUpdater struct {
	cfg *GeoIPServiceConfig
}

func (m *GeoIPServiceUpdater) UpdateIfRequired() error {
	isUpdateRequired, err := m.isUpdateRequired()
	if err != nil {
		return err
	}

	if isUpdateRequired {
		if err := m.Update(); err != nil {
			return err
		}
	}

	isDatabaseValid, err := m.validateDatabase()
	if err != nil {
		return err
	}

	if !isDatabaseValid {
		return fmt.Errorf("database checksum mismatch: expected %s", m.cfg.SHA1)
	}

	return nil
}

func (m *GeoIPServiceUpdater) Update() error {
	response, err := http.Get(m.cfg.URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	database, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(m.cfg.Path), 0444); err != nil {
		return err
	}

	if err := ioutil.WriteFile(m.cfg.Path, database, 0444); err != nil {
		return err
	}

	return nil
}

func (m *GeoIPServiceUpdater) isUpdateRequired() (bool, error) {
	fileInfo, err := os.Stat(m.cfg.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}

		return false, err
	}

	if fileInfo.IsDir() {
		return false, fmt.Errorf(`%s" is a directory`, m.cfg.Path)
	}

	isDatabaseValid, err := m.validateDatabase()
	if err != nil {
		return false, err
	}

	return !isDatabaseValid, nil
}

// ValidateDatabase reads the geo IP database and validates its checksum.
// The error returned signals about I/O error.
func (m *GeoIPServiceUpdater) validateDatabase() (bool, error) {
	database, err := ioutil.ReadFile(m.cfg.Path)
	if err != nil {
		return false, err
	}

	expectedHash, err := hex.DecodeString(m.cfg.SHA1)
	if err != nil {
		return false, err
	}

	hash := sha1.Sum(database)

	return bytes.Equal(hash[:], expectedHash), nil
}

type GeoIPService struct {
	updater  *GeoIPServiceUpdater
	database *geoip2.Reader
}

// NewGeoIPService constructs a new geo IP service.
func NewGeoIPService(cfg *GeoIPServiceConfig) (*GeoIPService, error) {
	if err := cfg.Normalize(); err != nil {
		return nil, err
	}

	updater := &GeoIPServiceUpdater{cfg: cfg}
	if err := updater.UpdateIfRequired(); err != nil {
		return nil, err
	}

	database, err := geoip2.Open(cfg.Path)
	if err != nil {
		return nil, err
	}

	m := &GeoIPService{
		updater:  updater,
		database: database,
	}

	return m, nil
}

func (m *GeoIPService) Country(addr net.IP) (*geoip2.Country, error) {
	return m.database.Country(addr)
}

func (m *GeoIPService) Close() error {
	multiErr := multierror.NewMultiError()

	if err := m.database.Close(); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	return multiErr.ErrorOrNil()
}
