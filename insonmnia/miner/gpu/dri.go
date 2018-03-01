package gpu

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

const (
	pciSlotName = "PCI_SLOT_NAME"
)

var (
	devDRICardNameRe = regexp.MustCompile(`^(card)([\d]+)`)
	fullDevicePathRe = regexp.MustCompile(`(\/dev\/[a-z]+)(\d+)`)

	errBadNameFormat = errors.New("bad name format")
)

type DRICard struct {
	Num      int
	Name     string
	Path     string
	Devices  []string
	DeviceID uint64
	VendorID uint64
	Major    uint64
	Minor    uint64
	PCIBusID string
}

func NewDRICard(num int, name, path string) (DRICard, error) {
	c := DRICard{
		Num:  num,
		Name: name,
		Path: path,
	}

	err := c.collectRelatedDevices()
	if err != nil {
		return DRICard{}, err
	}

	err = c.collectDeviceVendorIDs()
	if err != nil {
		return DRICard{}, err
	}

	err = c.collectPCIBusID()
	if err != nil {
		return DRICard{}, err
	}

	return c, nil
}

// collectRelatedDevices collects related control and render Devices for card
func (card *DRICard) collectRelatedDevices() error {
	stat := syscall.Stat_t{}
	err := syscall.Stat(card.Path, &stat)
	if err != nil {
		return fmt.Errorf("cannot get device major:minor numbers: %v", err)
	}

	major := uint64(stat.Rdev / 256)
	minor := uint64(stat.Rdev % 256)

	// query /sys for related DRI Devices
	sysDevPath := fmt.Sprintf("/sys/dev/char/%d:%d/device/drm/", major, minor)

	card.Major = major
	card.Minor = minor

	fi, err := ioutil.ReadDir(sysDevPath)
	if err != nil {
		return err
	}

	var devices []string
	// add found Devices as part of the DRI
	for _, f := range fi {
		devices = append(devices, path.Join("/dev/dri/", f.Name()))
	}

	card.Devices = devices

	return nil
}

// collectDeviceVendorIDs read vendor and device IDs from /sys and parses its values
func (card *DRICard) collectDeviceVendorIDs() error {
	vendorIDFile := fmt.Sprintf("/sys/class/drm/%s/device/vendor", card.Name)
	deviceIDFile := fmt.Sprintf("/sys/class/drm/%s/device/device", card.Name)

	vendorID, err := readSysClassValue(vendorIDFile)
	if err != nil {
		return err
	}

	deviceID, err := readSysClassValue(deviceIDFile)
	if err != nil {
		return err
	}

	card.DeviceID = deviceID
	card.VendorID = vendorID

	return nil
}

func (card *DRICard) collectPCIBusID() error {
	p := fmt.Sprintf("/sys/class/drm/%s/device/uevent", card.Name)

	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		id, ok := parsePCISlotName(line)
		if ok {
			card.PCIBusID = id
			return nil
		}
	}

	return errors.New("cannot find PCI slot name into the file")
}

// readSysClassValue reads value from /sys/class/xxx file and
// try to parse it as integer
func readSysClassValue(f string) (uint64, error) {
	raw, err := ioutil.ReadFile(f)
	if err != nil {
		return 0, err
	}

	return parseSysClassValue(raw)
}

// parseSysClassValue parses data that was read from /sys/class/xxx as uint64 value
func parseSysClassValue(v []byte) (uint64, error) {
	hexWithEOL := strings.Replace(string(v), "0x", "", 1)
	hex := strings.TrimSpace(hexWithEOL)
	return strconv.ParseUint(hex, 16, 64)
}

// parsePCISlotName receive value from the uevent file
// and return PCI slot name if any
func parsePCISlotName(s string) (string, bool) {
	if strings.Contains(s, pciSlotName) {
		parts := strings.Split(s, "=")
		if len(parts) == 2 {
			return parts[1], true
		}
	}

	return "", false
}

// CollectDRICardDevices traverses overs /dev/dri and collect card Devices
// which can be bound into the container
func CollectDRICardDevices() ([]DRICard, error) {
	var cards []DRICard

	ls, err := ioutil.ReadDir("/dev/dri")
	if err != nil {
		return cards, err
	}

	for _, ff := range ls {
		c, err := newCardDeviceByName(ff.Name())
		if err != nil {
			if err == errBadNameFormat {
				continue
			}

			return []DRICard{}, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

// newCardDeviceByName returns DRI device by given name.
func newCardDeviceByName(s string) (DRICard, error) {
	m := devDRICardNameRe.FindStringSubmatch(s)
	match := m != nil && len(m) == 3

	if !match {
		return DRICard{}, errBadNameFormat
	}

	// do not check error because regexp matches by this group by numeric value
	v, _ := strconv.ParseInt(m[2], 10, 64)

	return NewDRICard(int(v), m[0], path.Join("/dev/dri", m[0]))
}

// newCardByDevicePath returns related DRI card device for
// given NVidia device path. Ex: /dev/nvidia1 -> /dev/dri/card1.
func newCardByDevicePath(s string) (DRICard, error) {
	m := fullDevicePathRe.FindStringSubmatch(s)

	if len(m) != 3 {
		return DRICard{}, fmt.Errorf("cannot extract device index from %s", s)
	}

	name := fmt.Sprintf("card%s", m[2])
	card, err := newCardDeviceByName(name)
	if err != nil {
		return DRICard{}, err
	}

	return card, nil
}
