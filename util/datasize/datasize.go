package datasize

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type bitSize uint64

type ByteSize struct {
	bitSize
}

type BitRate struct {
	bitSize
}

const (
	B   bitSize = 8
	KiB         = B << 10
	MiB         = KiB << 10
	GiB         = MiB << 10
	TiB         = GiB << 10
	PiB         = TiB << 10
	EiB         = PiB << 10

	KB = B * 1000
	MB = KB * 1000
	GB = MB * 1000
	TB = GB * 1000
	PB = TB * 1000
	EB = PB * 1000

	bit   bitSize = 1
	Kibit         = bit << 10
	Mibit         = Kibit << 10
	Gibit         = Mibit << 10
	Tibit         = Gibit << 10
	Pibit         = Tibit << 10
	Eibit         = Pibit << 10

	kbit = bit * 1000
	Mbit = kbit * 1000
	Gbit = Mbit * 1000
	Tbit = Gbit * 1000
	Pbit = Tbit * 1000
	Ebit = Pbit * 1000

	bitDimFlag     = 0x1
	byteDimFlag    = 0x2
	decimalDimFlag = 0x4
	binaryDimFlag  = 0x8

	fnUnmarshalText string = "UnmarshalText"
	maxUint64       uint64 = (1 << 64) - 1
	cutoff          uint64 = maxUint64 / 10
)

type Dimension struct {
	spelling string
	size     bitSize
	flags    uint8
}

// Suffixes are intentionally sorted by size
var dimensions = []Dimension{
	{"EiB", EiB, byteDimFlag | binaryDimFlag},
	{"EB", EB, byteDimFlag | decimalDimFlag},
	{"Eibit", Eibit, bitDimFlag | binaryDimFlag},
	{"Eib", Eibit, bitDimFlag | binaryDimFlag},
	{"Ebit", Ebit, bitDimFlag | decimalDimFlag},
	{"Eb", Ebit, bitDimFlag | decimalDimFlag},

	{"PiB", PiB, byteDimFlag | binaryDimFlag},
	{"PB", PB, byteDimFlag | decimalDimFlag},
	{"Pibit", Pibit, bitDimFlag | binaryDimFlag},
	{"Pib", Pibit, bitDimFlag | binaryDimFlag},
	{"Pbit", Pbit, bitDimFlag | decimalDimFlag},
	{"Pb", Pbit, bitDimFlag | decimalDimFlag},

	{"TiB", TiB, byteDimFlag | binaryDimFlag},
	{"TB", TB, byteDimFlag | decimalDimFlag},
	{"Tibit", Tibit, bitDimFlag | binaryDimFlag},
	{"Tib", Tibit, bitDimFlag | binaryDimFlag},
	{"Tbit", Tbit, bitDimFlag | decimalDimFlag},
	{"Tb", Tbit, bitDimFlag | decimalDimFlag},

	{"GiB", GiB, byteDimFlag | binaryDimFlag},
	{"GB", GB, byteDimFlag | decimalDimFlag},
	{"Gibit", Gibit, bitDimFlag | binaryDimFlag},
	{"Gib", Gibit, bitDimFlag | binaryDimFlag},
	{"Gbit", Gbit, bitDimFlag | decimalDimFlag},
	{"Gb", Gbit, bitDimFlag | decimalDimFlag},

	{"MiB", MiB, byteDimFlag | binaryDimFlag},
	{"MB", MB, byteDimFlag | decimalDimFlag},
	{"Mibit", Mibit, bitDimFlag | binaryDimFlag},
	{"Mib", Mibit, bitDimFlag | binaryDimFlag},
	{"Mbit", Mbit, bitDimFlag | decimalDimFlag},
	{"Mb", Mbit, bitDimFlag | decimalDimFlag},

	{"KiB", KiB, byteDimFlag | binaryDimFlag},
	{"kB", KB, byteDimFlag | decimalDimFlag},
	{"KB", KB, byteDimFlag | decimalDimFlag},
	{"Kibit", Kibit, bitDimFlag | binaryDimFlag},
	{"Kib", Kibit, bitDimFlag | binaryDimFlag},
	{"Kbit", kbit, bitDimFlag | decimalDimFlag},
	{"kbit", kbit, bitDimFlag | decimalDimFlag},
	{"Kb", kbit, bitDimFlag | decimalDimFlag},

	{"B", B, byteDimFlag | decimalDimFlag},

	{"bit", bit, bitDimFlag | decimalDimFlag},
}

func (b bitSize) dimensionCount(dimension bitSize) float64 {
	v := b / dimension
	r := b % dimension
	return float64(v) + float64(r)/float64(dimension)
}

func (b bitSize) Bits() uint64   { return uint64(b) }
func (b bitSize) Bytes() float64 { return b.dimensionCount(B) }

func (b bitSize) KBits() float64   { return b.dimensionCount(kbit) }
func (b bitSize) KBytes() float64  { return b.dimensionCount(KB) }
func (b bitSize) Kibits() float64  { return b.dimensionCount(Kibit) }
func (b bitSize) KiBytes() float64 { return b.dimensionCount(KiB) }

func (b bitSize) Mbits() float64   { return b.dimensionCount(Mbit) }
func (b bitSize) MBytes() float64  { return b.dimensionCount(MB) }
func (b bitSize) Mibits() float64  { return b.dimensionCount(Mibit) }
func (b bitSize) MiBytes() float64 { return b.dimensionCount(MiB) }

func (b bitSize) GBits() float64   { return b.dimensionCount(Gbit) }
func (b bitSize) GBytes() float64  { return b.dimensionCount(GB) }
func (b bitSize) GiBits() float64  { return b.dimensionCount(Gibit) }
func (b bitSize) GiBytes() float64 { return b.dimensionCount(GiB) }

func (b bitSize) TBits() float64   { return b.dimensionCount(Tbit) }
func (b bitSize) TBytes() float64  { return b.dimensionCount(TB) }
func (b bitSize) TiBits() float64  { return b.dimensionCount(Tibit) }
func (b bitSize) TiBytes() float64 { return b.dimensionCount(TiB) }

func (b bitSize) PBits() float64   { return b.dimensionCount(Pbit) }
func (b bitSize) PBytes() float64  { return b.dimensionCount(PB) }
func (b bitSize) PiBits() float64  { return b.dimensionCount(Pibit) }
func (b bitSize) PiBytes() float64 { return b.dimensionCount(PiB) }

func (b bitSize) EBits() float64   { return b.dimensionCount(Ebit) }
func (b bitSize) EBytes() float64  { return b.dimensionCount(EB) }
func (b bitSize) EiBits() float64  { return b.dimensionCount(Eibit) }
func (b bitSize) EiBytes() float64 { return b.dimensionCount(EiB) }

func (b bitSize) HumanReadableString(flags uint8) string {
	if b == 0 {
		return "0B"
	}
	for _, dimension := range dimensions {
		if dimension.flags&flags == flags && b > dimension.size {
			return fmt.Sprintf("%.3f %s", b.dimensionCount(dimension.size), dimension.spelling)
		}
	}
	return fmt.Sprintf("%d bit", b)
}

func (b bitSize) PreciseString(flags uint8) string {
	if b == 0 {
		return "0B"
	}
	for _, dimension := range dimensions {
		if dimension.flags&flags == flags && b%dimension.size == 0 {
			return fmt.Sprintf("%d %s", b/dimension.size, dimension.spelling)
		}
	}
	return fmt.Sprintf("%d bit", b)
}

func splitDimension(text []byte) (dimension string, size interface{}, err error) {
	str := string(text)

	parts := strings.FieldsFunc(str, func(c rune) bool {
		return c == ' '
	})

	var sizeStr string

	//TODO: Is it ok? empty string is being parsed as default(0) value.
	if len(parts) == 0 {
		return "", uint64(0), nil
	}
	if len(parts) == 1 {
		dimension = strings.TrimLeftFunc(parts[0], func(r rune) bool {
			return '0' <= r && r <= '9'
		})
		sizeStr = strings.TrimRightFunc(parts[0], func(r rune) bool {
			return '0' > r || r > '9'
		})
	} else {
		if len(parts) != 2 {
			err = fmt.Errorf(`could not parse bitSize - "%s" can not be split to 2 parts`, str)
			return
		}
		dimension = parts[1]
		sizeStr = parts[0]
	}

	size, err = strconv.ParseUint(sizeStr, 10, 64)
	if err != nil {
		size, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			err = fmt.Errorf("could not parse bitSize numeric part -  %s", err)
		}
	}
	return
}

func getPossibleDimensions(flags uint8) []string {
	result := []string{}
	for _, dim := range dimensions {
		if dim.flags&flags == flags {
			result = append(result, dim.spelling)
		}
	}
	return result
}

func (b *bitSize) Unmarshal(size interface{}, dimension string, flags uint8) error {
	for _, dim := range dimensions {
		if dim.flags&flags == flags && dim.spelling == dimension {
			// Overflow check
			switch size := size.(type) {
			case float64:
				if float64(^uint64(0)) < size*float64(dim.size) {
					return errors.New("could not parse bitSize - too big value")
				}
				*b = bitSize(size * float64(dim.size))
				return nil
			case uint64:
				if size != 0 && ^uint64(0)/size < uint64(dim.size) {
					return errors.New("could not parse bitSize - too big value")
				}
				*b = bitSize(size) * dim.size
				return nil
			default:
				return errors.New("invalid size for unmarshall, should be float64 or uint64")
			}
		}
	}
	possibleDimensions := "[" + strings.Join(getPossibleDimensions(flags), ",") + "]"
	return fmt.Errorf(`could not parse bitSize - invalid suffix "%s", possible values are %s`,
		dimension, possibleDimensions)
}

func (d *BitRate) UnmarshalText(text []byte) error {
	dimension, size, err := splitDimension(text)
	if err != nil {
		return err
	}
	//TODO: Do we really want this kind of behaviour?
	if len(dimension) == 0 {
		dimension = "bit/s"
	}
	if !strings.HasSuffix(dimension, "/s") {
		return fmt.Errorf(`unknown data rate dimension "%s", dimension should have "/s" suffix`, dimension)
	}
	dimension = strings.TrimSuffix(dimension, "/s")
	//TODO: is this ok to parse BitRate in bytes per second?
	return d.bitSize.Unmarshal(size, dimension, 0)
}

func (d BitRate) MarshalText() ([]byte, error) {
	str := d.PreciseString(bitDimFlag) + "/s"
	return []byte(str), nil
}

func (d BitRate) HumanReadable() string {
	return d.HumanReadableBin()
}

func (d BitRate) HumanReadableDec() string {
	return d.HumanReadableString(bitDimFlag|decimalDimFlag) + "/s"
}

func (d BitRate) HumanReadableBin() string {
	return d.HumanReadableString(bitDimFlag|binaryDimFlag) + "/s"
}

func (d *ByteSize) UnmarshalText(text []byte) error {
	dimension, size, err := splitDimension(text)
	if err != nil {
		return err
	}
	if len(dimension) == 0 {
		dimension = "B"
	}
	return d.bitSize.Unmarshal(size, dimension, byteDimFlag)
}

func (d ByteSize) MarshalText() ([]byte, error) {
	str := d.PreciseString(byteDimFlag)
	return []byte(str), nil
}

func (d ByteSize) HumanReadable() string {
	return d.HumanReadableBin()
}

func (d ByteSize) HumanReadableDec() string {
	return d.HumanReadableString(byteDimFlag | decimalDimFlag)
}

func (d ByteSize) HumanReadableBin() string {
	return d.HumanReadableString(byteDimFlag | binaryDimFlag)
}

func (b ByteSize) Bytes() uint64 {
	return uint64(b.bitSize / B)
}

func NewByteSize(bytes uint64) ByteSize {
	return ByteSize{bitSize(bytes) * B}
}

func NewBitRate(bitsPerSec uint64) BitRate {
	return BitRate{bitSize(bitsPerSec)}
}
