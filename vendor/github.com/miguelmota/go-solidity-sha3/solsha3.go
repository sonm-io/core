package solsha3

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

// Address address
func Address(input interface{}) []byte {
	switch v := input.(type) {
	case common.Address:
		return v.Bytes()[:]
	case string:
		return common.HexToAddress(v).Bytes()[:]
	default:
		return common.HexToAddress("").Bytes()[:]
	}
}

// Uint256 uint256
func Uint256(input interface{}) []byte {
	switch v := input.(type) {
	case *big.Int:
		return abi.U256(v)
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		return abi.U256(bn)
	default:
		return common.RightPadBytes([]byte(""), 32)
	}
}

// Uint128 uint128
func Uint128(input interface{}) []byte {
	switch v := input.(type) {
	case *big.Int:
		return common.LeftPadBytes(v.Bytes(), 16)
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		return common.LeftPadBytes(bn.Bytes(), 16)
	default:
		return common.LeftPadBytes([]byte(""), 16)
	}
}

// Uint64 uint64
func Uint64(input interface{}) []byte {
	b := new(bytes.Buffer)
	switch v := input.(type) {
	case *big.Int:
		binary.Write(b, binary.BigEndian, v.Uint64())
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.Write(b, binary.BigEndian, bn.Uint64())
	case uint64:
		binary.Write(b, binary.BigEndian, v)
	case uint32:
		binary.Write(b, binary.BigEndian, uint64(v))
	case uint16:
		binary.Write(b, binary.BigEndian, uint64(v))
	case uint8:
		binary.Write(b, binary.BigEndian, uint64(v))
	case uint:
		binary.Write(b, binary.BigEndian, uint64(v))
	case int64:
		binary.Write(b, binary.BigEndian, uint64(v))
	case int32:
		binary.Write(b, binary.BigEndian, uint64(v))
	case int16:
		binary.Write(b, binary.BigEndian, uint64(v))
	case int8:
		binary.Write(b, binary.BigEndian, uint64(v))
	case int:
		binary.Write(b, binary.BigEndian, uint64(v))
	default:
		binary.Write(b, binary.BigEndian, uint64(0))
	}
	return b.Bytes()
}

// Uint32 uint32
func Uint32(input interface{}) []byte {
	b := new(bytes.Buffer)
	switch v := input.(type) {
	case *big.Int:
		binary.Write(b, binary.BigEndian, uint32(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.Write(b, binary.BigEndian, uint32(bn.Uint64()))
	case uint64:
		binary.Write(b, binary.BigEndian, uint32(v))
	case uint32:
		binary.Write(b, binary.BigEndian, uint32(v))
	case uint16:
		binary.Write(b, binary.BigEndian, uint32(v))
	case uint8:
		binary.Write(b, binary.BigEndian, uint32(v))
	case uint:
		binary.Write(b, binary.BigEndian, uint32(v))
	case int64:
		binary.Write(b, binary.BigEndian, uint32(v))
	case int32:
		binary.Write(b, binary.BigEndian, v)
	case int16:
		binary.Write(b, binary.BigEndian, uint32(v))
	case int8:
		binary.Write(b, binary.BigEndian, uint32(v))
	case int:
		binary.Write(b, binary.BigEndian, uint32(v))
	default:
		binary.Write(b, binary.BigEndian, uint32(0))
	}
	return b.Bytes()
}

// Uint16 uint16
func Uint16(input interface{}) []byte {
	b := new(bytes.Buffer)
	switch v := input.(type) {
	case *big.Int:
		binary.Write(b, binary.BigEndian, uint16(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.Write(b, binary.BigEndian, uint16(bn.Uint64()))
	case uint64:
		binary.Write(b, binary.BigEndian, uint16(v))
	case uint32:
		binary.Write(b, binary.BigEndian, uint16(v))
	case uint16:
		binary.Write(b, binary.BigEndian, v)
	case uint8:
		binary.Write(b, binary.BigEndian, uint16(v))
	case uint:
		binary.Write(b, binary.BigEndian, uint16(v))
	case int64:
		binary.Write(b, binary.BigEndian, uint16(v))
	case int32:
		binary.Write(b, binary.BigEndian, uint16(v))
	case int16:
		binary.Write(b, binary.BigEndian, uint16(v))
	case int8:
		binary.Write(b, binary.BigEndian, uint16(v))
	case int:
		binary.Write(b, binary.BigEndian, uint16(v))
	default:
		binary.Write(b, binary.BigEndian, uint16(0))
	}
	return b.Bytes()
}

// Uint8 uint8
func Uint8(input interface{}) []byte {
	b := new(bytes.Buffer)
	switch v := input.(type) {
	case *big.Int:
		binary.Write(b, binary.BigEndian, uint8(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.Write(b, binary.BigEndian, uint8(bn.Uint64()))
	case uint64:
		binary.Write(b, binary.BigEndian, uint8(v))
	case uint32:
		binary.Write(b, binary.BigEndian, uint8(v))
	case uint16:
		binary.Write(b, binary.BigEndian, uint8(v))
	case uint8:
		binary.Write(b, binary.BigEndian, v)
	case uint:
		binary.Write(b, binary.BigEndian, uint8(v))
	case int64:
		binary.Write(b, binary.BigEndian, uint8(v))
	case int32:
		binary.Write(b, binary.BigEndian, uint8(v))
	case int16:
		binary.Write(b, binary.BigEndian, uint8(v))
	case int8:
		binary.Write(b, binary.BigEndian, uint8(v))
	case int:
		binary.Write(b, binary.BigEndian, uint8(v))
	default:
		binary.Write(b, binary.BigEndian, uint8(0))
	}
	return b.Bytes()
}

// Int256 int256
func Int256(input interface{}) []byte {
	switch v := input.(type) {
	case *big.Int:
		return common.LeftPadBytes(v.Bytes(), 32)
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		return common.LeftPadBytes(bn.Bytes(), 32)
	case uint64:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case uint32:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case uint16:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case uint8:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case uint:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case int64:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case int32:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case int16:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case int8:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	case int:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 32)
	default:
		bn := big.NewInt(int64(0))
		return common.LeftPadBytes(bn.Bytes(), 32)
	}
}

// Int128 int128
func Int128(input interface{}) []byte {
	switch v := input.(type) {
	case *big.Int:
		return common.LeftPadBytes(v.Bytes(), 16)
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		return common.LeftPadBytes(bn.Bytes(), 16)
	case uint64:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case uint32:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case uint16:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case uint8:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case uint:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case int64:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case int32:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case int16:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case int8:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	case int:
		bn := big.NewInt(int64(v))
		return common.LeftPadBytes(bn.Bytes(), 16)
	default:
		bn := big.NewInt(int64(0))
		return common.LeftPadBytes(bn.Bytes(), 16)
	}
}

// Int64 int64
func Int64(input interface{}) []byte {
	b := make([]byte, 8)
	switch v := input.(type) {
	case *big.Int:
		binary.BigEndian.PutUint64(b, v.Uint64())
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.BigEndian.PutUint64(b, bn.Uint64())
	case uint64:
		binary.BigEndian.PutUint64(b, v)
	case uint32:
		binary.BigEndian.PutUint64(b, uint64(v))
	case uint16:
		binary.BigEndian.PutUint64(b, uint64(v))
	case uint8:
		binary.BigEndian.PutUint64(b, uint64(v))
	case uint:
		binary.BigEndian.PutUint64(b, uint64(v))
	case int64:
		binary.BigEndian.PutUint64(b, uint64(v))
	case int32:
		binary.BigEndian.PutUint64(b, uint64(v))
	case int16:
		binary.BigEndian.PutUint64(b, uint64(v))
	case int8:
		binary.BigEndian.PutUint64(b, uint64(v))
	case int:
		binary.BigEndian.PutUint64(b, uint64(v))
	default:
		binary.BigEndian.PutUint64(b, uint64(0))
	}
	return b
}

// Int32 int32
func Int32(input interface{}) []byte {
	b := make([]byte, 4)
	switch v := input.(type) {
	case *big.Int:
		binary.BigEndian.PutUint32(b, uint32(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.BigEndian.PutUint32(b, uint32(bn.Uint64()))
	case uint64:
		binary.BigEndian.PutUint32(b, uint32(v))
	case uint32:
		binary.BigEndian.PutUint32(b, v)
	case uint16:
		binary.BigEndian.PutUint32(b, uint32(v))
	case uint8:
		binary.BigEndian.PutUint32(b, uint32(v))
	case uint:
		binary.BigEndian.PutUint32(b, uint32(v))
	case int64:
		binary.BigEndian.PutUint32(b, uint32(v))
	case int32:
		binary.BigEndian.PutUint32(b, uint32(v))
	case int16:
		binary.BigEndian.PutUint32(b, uint32(v))
	case int8:
		binary.BigEndian.PutUint32(b, uint32(v))
	case int:
		binary.BigEndian.PutUint32(b, uint32(v))
	default:
		binary.BigEndian.PutUint32(b, uint32(0))
	}
	return b
}

// Int16 int16
func Int16(input interface{}) []byte {
	b := make([]byte, 2)
	switch v := input.(type) {
	case *big.Int:
		binary.BigEndian.PutUint16(b, uint16(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		binary.BigEndian.PutUint16(b, uint16(bn.Uint64()))
	case uint64:
		binary.BigEndian.PutUint16(b, uint16(v))
	case uint32:
		binary.BigEndian.PutUint16(b, uint16(v))
	case uint16:
		binary.BigEndian.PutUint16(b, v)
	case uint8:
		binary.BigEndian.PutUint16(b, uint16(v))
	case uint:
		binary.BigEndian.PutUint16(b, uint16(v))
	case int64:
		binary.BigEndian.PutUint16(b, uint16(v))
	case int32:
		binary.BigEndian.PutUint16(b, uint16(v))
	case int16:
		binary.BigEndian.PutUint16(b, uint16(v))
	case int8:
		binary.BigEndian.PutUint16(b, uint16(v))
	case int:
		binary.BigEndian.PutUint16(b, uint16(v))
	default:
		binary.BigEndian.PutUint16(b, uint16(0))
	}
	return b
}

// Int8 int8
func Int8(input interface{}) []byte {
	b := make([]byte, 1)
	switch v := input.(type) {
	case *big.Int:
		b[0] = byte(int8(v.Uint64()))
	case string:
		bn := new(big.Int)
		bn.SetString(v, 10)
		b[0] = byte(int8(bn.Uint64()))
	case uint64:
		b[0] = byte(int8(v))
	case uint32:
		b[0] = byte(int8(v))
	case uint16:
		b[0] = byte(int8(v))
	case uint8:
		b[0] = byte(int8(v))
	case uint:
		b[0] = byte(int8(v))
	case int64:
		b[0] = byte(int8(v))
	case int32:
		b[0] = byte(int8(v))
	case int16:
		b[0] = byte(int8(v))
	case int8:
		b[0] = byte(v)
	case int:
		b[0] = byte(int8(v))
	default:
		b[0] = byte(int8(0))
	}
	return b
}

// Bytes32 bytes32
func Bytes32(input interface{}) []byte {
	switch v := input.(type) {
	case [32]byte:
		return common.RightPadBytes(v[:], 32)
	case []byte:
		return common.RightPadBytes(v, 32)
	case string:
		str := fmt.Sprintf("%x", v)
		hexb, _ := hex.DecodeString(str)
		return common.RightPadBytes(hexb, 32)
	default:
		return common.RightPadBytes([]byte(""), 32)
	}
}

// String string
func String(input interface{}) []byte {
	switch v := input.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	default:
		return []byte("")
	}
}

// Bool bool
func Bool(input interface{}) []byte {
	switch v := input.(type) {
	case bool:
		if v {
			return []byte{0x1}
		}
		return []byte{0x0}
	default:
		return []byte{0x0}
	}
}

// ConcatByteSlices concat byte slices
func ConcatByteSlices(arrays ...[]byte) []byte {
	var result []byte

	for _, b := range arrays {
		result = append(result, b...)
	}

	return result
}

// SoliditySHA3 solidity sha3
func SoliditySHA3(data ...[]byte) []byte {
	var result []byte

	hash := sha3.NewKeccak256()
	bs := ConcatByteSlices(data...)

	hash.Write(bs)
	result = hash.Sum(result)

	return result
}

// SoliditySHA3WithPrefix solidity sha3 with prefix
func SoliditySHA3WithPrefix(data []byte) []byte {
	result := SoliditySHA3(
		ConcatByteSlices(
			[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(data))),
			data,
		),
	)

	return result
}
