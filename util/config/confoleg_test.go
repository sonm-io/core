package config

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestLoadIntoStruct(t *testing.T) {
	type T struct {
		Name string
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
}

func TestErrLoadDecodeFailed(t *testing.T) {
	type T struct {
		Name string
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan\n  (): abc")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "yaml: line 2: mapping values are not allowed in this context")
}

func TestLoadIntoStructWithInt(t *testing.T) {
	type T struct {
		Name string
		Age  int
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan\nAge: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
	assert.Equal(t, 42, cfg.Age)
}

func TestLoadIntoStructWithInt8(t *testing.T) {
	type T struct {
		Age int8
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, int8(42), cfg.Age)
}

func TestLoadIntoStructWithInt16(t *testing.T) {
	type T struct {
		Age int16
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, int16(42), cfg.Age)
}

func TestLoadIntoStructWithInt32(t *testing.T) {
	type T struct {
		Age int32
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, int32(42), cfg.Age)
}

func TestLoadIntoStructWithInt64(t *testing.T) {
	type T struct {
		Age int64
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, int64(42), cfg.Age)
}

func TestLoadIntoStructWithInt64Large(t *testing.T) {
	type T struct {
		Age int64
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 9223372036854775807")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, int64(9223372036854775807), cfg.Age)
}

func TestErrLoadIntoStructWithInt8Overflow(t *testing.T) {
	type T struct {
		Age int8
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 420")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Age` expected type `int8`, but actual value 420 overflows it")
}

func TestErrLoadIntoStructWithInt8Underflow(t *testing.T) {
	type T struct {
		Age int8
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: -420")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Age` expected type `int8`, but actual value -420 overflows it")
}

func TestLoadIntoStructWithUInt8(t *testing.T) {
	type T struct {
		Age uint8
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, uint8(42), cfg.Age)
}

func TestLoadIntoStructWithUInt16(t *testing.T) {
	type T struct {
		Age uint16
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, uint16(42), cfg.Age)
}

func TestLoadIntoStructWithUInt32(t *testing.T) {
	type T struct {
		Age uint32
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, uint32(42), cfg.Age)
}

func TestLoadIntoStructWithUInt64(t *testing.T) {
	type T struct {
		Age uint64
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, uint64(42), cfg.Age)
}

func TestErrLoadIntoStructWithUint8Overflow(t *testing.T) {
	type T struct {
		Age uint8
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 256")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Age` expected type `uint8`, but actual value 256 overflows it")
}

func TestErrLoadIntoStructWithUint8Underflow(t *testing.T) {
	type T struct {
		Age uint8
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: -42")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Age` expected type `uint8`, but actual value -42 overflows it")
}

func TestLoadIntoStructWithFloat32(t *testing.T) {
	type T struct {
		Age float32
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 42.1415")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, float32(42.1415), cfg.Age)
}

func TestLoadIntoStructWithFloat64(t *testing.T) {
	type T struct {
		Age float64
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Age: 42.1415")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, 42.1415, cfg.Age)
}

func TestLoadIntoStructWithBool(t *testing.T) {
	type T struct {
		Bad bool
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Bad: true")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, true, cfg.Bad)
}

func TestLoadIntoStructFromLowercase(t *testing.T) {
	type T struct {
		Name string
		Age  int
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("name: Ivan\nage: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
	assert.Equal(t, 42, cfg.Age)
}

func TestLoadIntoStructFromSnakeCase(t *testing.T) {
	type T struct {
		Name   string
		MaxAge int
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("name: Ivan\nmax_age: 65")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
	assert.Equal(t, 65, cfg.MaxAge)
}

func TestLoadIntoStructRenamed(t *testing.T) {
	type T struct {
		Name string `yaml:"nickname"`
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("nickname: Ivan")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
}

func TestLoadIntoStructWithIntInsteadOfString(t *testing.T) {
	type T struct {
		Name string
		Age  string
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan\nAge: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "42", cfg.Age)
}

func TestErrLoadIntoStructWithBoolInsteadOfInt(t *testing.T) {
	type T struct {
		Name string
		Age  int
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan\nAge: true")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Age` expected type `int`, but actual type is `bool`")
}

func TestErrLoadIntoStructWithFloat64InsteadOfInt8(t *testing.T) {
	type T struct {
		Name string
		Age  int8
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan\nAge: 0.42")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Age` expected type `int8`, but actual type is `float64`")
}

func TestErrLoadIntoStructWithBoolInsteadOfUint64(t *testing.T) {
	type T struct {
		Name string
		Age  uint64
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan\nAge: true")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Age` expected type `uint64`, but actual type is `bool`")
}

func TestLoadIntoNestedStruct(t *testing.T) {
	type Head struct {
		BrainID int
	}
	type Cat struct {
		Name string
		Head Head
	}

	cfg := Cat{}
	err := Load(bytes.NewReader([]byte("name: Barsik\nhead:\n  brain_id: 42")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Barsik", cfg.Name)
	assert.Equal(t, 42, cfg.Head.BrainID)
}

func TestErrLoadIntoNestedStructTypeMismatch(t *testing.T) {
	type Head struct {
		BrainID int
	}
	type Cat struct {
		Name string
		Head Head
	}

	cfg := Cat{}
	err := Load(bytes.NewReader([]byte("name: Barsik\nhead:\n  brain_id: true")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Head.BrainID` expected type `int`, but actual type is `bool`")
}

func TestLoadIntoStructWithArray(t *testing.T) {
	type Cat struct {
		Name string
		Age  int
	}

	type Cemetery struct {
		Cats []Cat
	}

	cfg := Cemetery{}
	err := Load(bytes.NewReader([]byte(`
cats:
  - name: Barsik
    age: 2
  - name: Jesus
    age: 33

`)), &cfg)

	require.NoError(t, err)
	require.Len(t, cfg.Cats, 2)
	assert.Equal(t, "Barsik", cfg.Cats[0].Name)
	assert.Equal(t, 2, cfg.Cats[0].Age)
	assert.Equal(t, "Jesus", cfg.Cats[1].Name)
	assert.Equal(t, 33, cfg.Cats[1].Age)
}

func TestErrLoadIntoStructWithArrayTypeMismatch(t *testing.T) {
	type Cat struct {
		Name string
		Age  int
	}

	type Cemetery struct {
		Cats []Cat
	}

	cfg := Cemetery{}
	err := Load(bytes.NewReader([]byte(`
cats:
  - name: Barsik
    age: 2
  - name: Jesus
    age: true

`)), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Cats.[1].Age` expected type `int`, but actual type is `bool`")
}

func TestErrLoadIntoStructWithBoolInsteadOfFloat64(t *testing.T) {
	type T struct {
		Name string
		Age  float64
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan\nAge: true")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Age` expected type `float64`, but actual type is `bool`")
}

func TestLoadIntoMap(t *testing.T) {
	type Cat struct {
		Age int
	}

	type Cemetery struct {
		Cats map[string]Cat
	}

	cfg := Cemetery{}
	err := Load(bytes.NewReader([]byte(`
cats:
  Barsik:
    age: 2
  Jesus:
    age: 33
`)), &cfg)

	require.NoError(t, err)
	require.Equal(t, 2, len(cfg.Cats))
	assert.Equal(t, 2, cfg.Cats["Barsik"].Age)
	assert.Equal(t, 33, cfg.Cats["Jesus"].Age)
}

func TestLoadIntoMapWithPtr(t *testing.T) {
	type Cat struct {
		Age int
	}

	type Cemetery struct {
		Cats map[string]*Cat
	}

	cfg := Cemetery{}
	err := Load(bytes.NewReader([]byte(`
cats:
  Barsik:
    age: 2
  Jesus:
    age: 33
`)), &cfg)

	require.NoError(t, err)
	require.Equal(t, 2, len(cfg.Cats))
	assert.Equal(t, 2, cfg.Cats["Barsik"].Age)
	assert.Equal(t, 33, cfg.Cats["Jesus"].Age)
}

func TestLoadIntoStructWithArrayPtr(t *testing.T) {
	type Cat struct {
		Name string
		Age  int
	}

	type Cemetery struct {
		Cats []*Cat
	}

	cfg := Cemetery{}
	err := Load(bytes.NewReader([]byte(`
cats:
  - name: Barsik
    age: 2
  - name: Jesus
    age: 33

`)), &cfg)

	require.NoError(t, err)
	require.Len(t, cfg.Cats, 2)
	assert.Equal(t, "Barsik", cfg.Cats[0].Name)
	assert.Equal(t, 2, cfg.Cats[0].Age)
	assert.Equal(t, "Jesus", cfg.Cats[1].Name)
	assert.Equal(t, 33, cfg.Cats[1].Age)
}

func TestLoadIntoStructWithFixedArrayPtr(t *testing.T) {
	type Cat struct {
		Name string
		Age  int
	}

	type Cemetery struct {
		Cats [2]*Cat
	}

	cfg := Cemetery{}
	err := Load(bytes.NewReader([]byte(`
cats:
  - name: Barsik
    age: 2
  - name: Jesus
    age: 33

`)), &cfg)

	require.NoError(t, err)
	require.Len(t, cfg.Cats, 2)
	require.NotNil(t, cfg.Cats[0])
	require.NotNil(t, cfg.Cats[1])
	assert.Equal(t, "Barsik", cfg.Cats[0].Name)
	assert.Equal(t, 2, cfg.Cats[0].Age)
	assert.Equal(t, "Jesus", cfg.Cats[1].Name)
	assert.Equal(t, 33, cfg.Cats[1].Age)
}

func TestErrLoadIntoStructWithFixedArrayPtrLengthMismatch(t *testing.T) {
	type Cat struct {
		Name string
		Age  int
	}

	type Cemetery struct {
		Cats [1]*Cat
	}

	cfg := Cemetery{}
	err := Load(bytes.NewReader([]byte(`
cats:
  - name: Barsik
    age: 2
  - name: Jesus
    age: 33

`)), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Cats` expected length `1`, but actual length is `2`")
}

func TestErrLoadIntoStructWithArrayPtrTypeMismatch(t *testing.T) {
	type Cat struct {
		Name string
		Age  int
	}

	type Cemetery struct {
		Cats []*Cat
	}

	cfg := Cemetery{}
	err := Load(bytes.NewReader([]byte(`
cats: oh-long-john
`)), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `Cats` expected to have type `[]*config.Cat`, but actual type is `string`")
}

func TestErrLoadIntoStructWithIntRequired(t *testing.T) {
	type T struct {
		Name string
		Age  int `required:"true"`
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan")), &cfg)

	require.Error(t, err)
	assert.EqualError(t, err, "failed to decode: `.Age` is required")
}

func TestErrLoadIntoStructWithIntNotRequired(t *testing.T) {
	type T struct {
		Name string
		Age  int
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
	assert.Equal(t, 0, cfg.Age)
}

func TestLoadIntoStructWithIntDefault(t *testing.T) {
	type T struct {
		Name string
		Age  int `default:"42"`
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
	assert.Equal(t, 42, cfg.Age)
}

func TestLoadIntoStructWithDuration(t *testing.T) {
	type T struct {
		Name     string
		Lifetime time.Duration
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("name: Ivan\nlifetime: 42m")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
	assert.Equal(t, 42*time.Minute, cfg.Lifetime)
}

func TestLoadIntoStructWithDurationDefault(t *testing.T) {
	type T struct {
		Name     string
		Lifetime time.Duration `default:"42m"`
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
	assert.Equal(t, 42*time.Minute, cfg.Lifetime)
}

type Color int

func (m *Color) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	if v == "red" {
		*m = 1
	}

	return nil
}

func TestLoadIntoStructWithCustomUnmarshaller(t *testing.T) {
	type T struct {
		Name  string
		Color Color
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("name: Ivan\ncolor: red")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
	assert.Equal(t, Color(1), cfg.Color)
}

func TestLoadIntoStructWithDefaultAndCustomUnmarshaller(t *testing.T) {
	type T struct {
		Name  string
		Color Color `default:"red"`
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("Name: Ivan")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, "Ivan", cfg.Name)
	assert.Equal(t, Color(1), cfg.Color)
}

type Env struct {
	Key   string
	Value string
}

func (m *Env) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v int
	if err := unmarshal(&v); err != nil {
		return err
	}

	switch v {
	case 0:
		m.Key = "foo"
		m.Value = "bar"
	case 1:
		m.Key = "moo"
		m.Value = "baz"
	default:
		return fmt.Errorf("invalid format")
	}

	return nil
}

func TestLoadComplexWithYAMLUnmarshaller(t *testing.T) {
	type T struct {
		Env []Env
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte(`
env:
  - 0
  - 1
`)), &cfg)

	require.NoError(t, err)
	require.Len(t, cfg.Env, 2)
	assert.Equal(t, "foo", cfg.Env[0].Key)
	assert.Equal(t, "bar", cfg.Env[0].Value)
	assert.Equal(t, "moo", cfg.Env[1].Key)
	assert.Equal(t, "baz", cfg.Env[1].Value)
}

func TestLoadLoggingConfig(t *testing.T) {
	type T struct {
		Logging logging.Config
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte("logging:\n  level: debug")), &cfg)

	require.NoError(t, err)
	assert.Equal(t, zapcore.DebugLevel, cfg.Logging.LogLevel().Zap())
}

func TestLoadFloat64ToString(t *testing.T) {
	type T struct {
		Options map[string]string
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte(`
options:
  version: 3.1
`)), &cfg)

	require.NoError(t, err)
	require.Contains(t, cfg.Options, "version")
	assert.Equal(t, "3.1", cfg.Options["version"])
}

func TestLoadCommonAddress(t *testing.T) {
	type T struct {
		Master common.Address
	}

	cfg := T{}
	err := Load(bytes.NewReader([]byte(`master: "0x0000000000000000000000000000000000000001"`)), &cfg)

	require.NoError(t, err)
	assert.Equal(t, common.HexToAddress("0x0000000000000000000000000000000000000001"), cfg.Master)
}
