package config

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type Validate interface {
	Validate() error
}

// Debugger is a helper trait used for internal decoder debugging.
type Debugger interface {
	Debugf(template string, args ...interface{})
}

type printfDebugger struct{}

func (printfDebugger) Debugf(template string, args ...interface{}) {
	fmt.Printf(template+"\n", args...)
}

type Option func(o *options)

type options struct {
	tags *decodeTags
	d    Debugger
}

func newOptions() *options {
	return &options{
		tags: newDefaultDecodeTags(),
		d:    zap.NewNop().Sugar(),
	}
}

func WithDebugger(d Debugger) Option {
	return func(o *options) {
		o.d = d
	}
}

func WithPrintfDebugger() Option {
	return WithDebugger(printfDebugger{})
}

func WithRenameTag(tag string) Option {
	return func(o *options) {
		o.tags.Rename = tag
	}
}

func WithRequiredTag(tag string) Option {
	return func(o *options) {
		o.tags.Required = tag
	}
}

func WithDefaultTag(tag string) Option {
	return func(o *options) {
		o.tags.Default = tag
	}
}

type decodeTags struct {
	Rename   string
	Required string
	Default  string
}

func newDefaultDecodeTags() *decodeTags {
	return &decodeTags{
		Rename:   "yaml",
		Required: "required",
		Default:  "default",
	}
}

type Decoder struct {
	v    interface{}
	path string
	opts *options
}

func NewDecoder(v interface{}, options ...Option) *Decoder {
	opts := newOptions()
	for _, o := range options {
		o(opts)
	}

	return &Decoder{
		v:    v,
		opts: opts,
	}
}

func (m *Decoder) Decode(target interface{}) error {
	return m.decode(reflect.ValueOf(target).Elem())
}

func (m *Decoder) addDecoder(v interface{}, path string) *Decoder {
	if len(m.path) != 0 {
		path = fmt.Sprintf("%s.%s", m.path, path)
	}

	return &Decoder{
		v:    v,
		path: path,
		opts: m.opts,
	}
}

func (m *Decoder) decode(target reflect.Value) error {
	m.opts.d.Debugf("[ ] visiting `%s` `%v`: %T -> `%v`: %v", m.path, m.v, m.v, target, target.Type())
	defer m.opts.d.Debugf("[x] visiting `%s` `%v`: %T -> %v: %v", m.path, m.v, m.v, target, target.Type())

	if m.v == nil {
		return nil
	}

	if !reflect.ValueOf(m.v).IsValid() {
		// If the data value is invalid, then we just set the value to be the
		// zero value.
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	// TODO: try to custom unmarshal.

	if u, ok := reflect.New(target.Type()).Interface().(encoding.TextUnmarshaler); ok {
		if data, ok := m.v.(string); ok {
			m.opts.d.Debugf("deserializing from text: %s", string(data))
			if err := u.UnmarshalText([]byte(data)); err != nil {
				return m.errUnmarshal(err)
			}

			m.opts.d.Debugf("after deserializing from text: `%v: %T` -> `%v: %T`", m.v, m.v, reflect.Indirect(reflect.ValueOf(u)).Interface(), reflect.Indirect(reflect.ValueOf(u)).Interface())
			m.v = reflect.Indirect(reflect.ValueOf(u)).Interface()
			target.Set(m.Value())
			return nil
		}
	}

	// YAML aware conversion from string.
	if u, ok := reflect.New(target.Type()).Interface().(yaml.Unmarshaler); ok {
		data, err := yaml.Marshal(m.v)
		if err != nil {
			return err
		}

		m.opts.d.Debugf("deserializing from YAML")
		if err := yaml.Unmarshal(data, u); err != nil {
			return m.errUnmarshal(err)
		}

		m.opts.d.Debugf("after deserializing from YAML: `%v: %T` -> `%v: %T`", m.v, m.v, reflect.Indirect(reflect.ValueOf(u)).Interface(), reflect.Indirect(reflect.ValueOf(u)).Interface())
		m.v = reflect.Indirect(reflect.ValueOf(u)).Interface()
		target.Set(m.Value())
		return nil
	}

	if reflect.ValueOf(m.v).Kind() == reflect.String {
		if target.Type() == reflect.ValueOf(time.Duration(0)).Type() {
			v := time.Duration(0)
			if err := yaml.Unmarshal([]byte(m.v.(string)), &v); err != nil {
				return m.errUnmarshal(err)
			}

			m.v = reflect.Indirect(reflect.ValueOf(v)).Interface()
			target.Set(m.Value())
			return nil
		}
	}

	switch target.Kind() {
	case reflect.Bool:
		if err := m.ensureType(target.Type()); err != nil {
			return err
		}
		target.Set(m.Value())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return m.decodeInt(target)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return m.decodeUint(target)
	case reflect.Float32, reflect.Float64:
		return m.decodeFloat(target)
	case reflect.String:
		return m.decodeString(target)
	case reflect.Array:
		return m.decodeArray(target)
	case reflect.Slice:
		return m.decodeSlice(target)
	case reflect.Map:
		return m.decodeMap(target)
	case reflect.Ptr:
		return m.decodePtr(target)
	case reflect.Struct:
		return m.decodeStruct(target)
	default:
		return fmt.Errorf("unimplemented: `%v`", target.Kind())
	}
	return nil
}

func (m *Decoder) decodeInt(target reflect.Value) error {
	sourceValue := reflect.ValueOf(m.v)
	switch sourceValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sourceValueInt := sourceValue.Int()
		if target.OverflowInt(sourceValueInt) {
			return m.errOverflow(target.Type().String(), sourceValueInt)
		}
		target.SetInt(sourceValueInt)
	default:
		return m.errTypeMismatch(target.Type().String())
	}

	return nil
}

func (m *Decoder) decodeUint(target reflect.Value) error {
	sourceValue := reflect.ValueOf(m.v)
	switch sourceValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sourceValueInt := sourceValue.Int()
		if sourceValueInt < 0 {
			return m.errOverflow(target.Type().String(), sourceValueInt)
		}

		sourceValueUint := uint64(sourceValueInt)
		if target.OverflowUint(sourceValueUint) {
			return m.errOverflow(target.Type().String(), sourceValueInt)
		}
		target.SetUint(sourceValueUint)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		sourceValueUint := sourceValue.Uint()
		if target.OverflowUint(sourceValueUint) {
			return m.errOverflow(target.Type().String(), sourceValueUint)
		}
		target.SetUint(sourceValueUint)
	default:
		return m.errTypeMismatch(target.Type().String())
	}

	return nil
}

func (m *Decoder) decodeFloat(target reflect.Value) error {
	sourceValue := reflect.ValueOf(m.v)
	switch sourceValue.Kind() {
	case reflect.Float32, reflect.Float64:
		target.SetFloat(sourceValue.Float())
	default:
		return m.errTypeMismatch(target.Type().String())
	}

	return nil
}

func (m *Decoder) decodeString(target reflect.Value) error {
	sourceValue := reflect.ValueOf(m.v)
	switch sourceValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetString(fmt.Sprintf("%d", sourceValue.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		target.SetString(fmt.Sprintf("%d", sourceValue.Uint()))
	case reflect.Float32, reflect.Float64:
		target.SetString(strconv.FormatFloat(sourceValue.Float(), 'f', -1, 64))
	case reflect.String:
		target.SetString(sourceValue.String())
	default:
		return m.errTypeMismatch(target.Type().String())
	}

	return nil
}

func (m *Decoder) decodeArray(target reflect.Value) error {
	sourceValue, err := m.SliceOf(target.Type())
	if err != nil {
		return err
	}

	if len(sourceValue) != target.Len() {
		return m.errLengthMismatch(target.Len(), len(sourceValue))
	}

	for id := 0; id < len(sourceValue); id++ {
		if err := m.addDecoder(sourceValue[id], fmt.Sprintf("[%d]", id)).decode(target.Index(id)); err != nil {
			return err
		}
	}

	return nil
}

func (m *Decoder) decodeSlice(target reflect.Value) error {
	sourceValue, err := m.SliceOf(target.Type())
	if err != nil {
		return err
	}

	targetSlice := reflect.MakeSlice(target.Type(), len(sourceValue), len(sourceValue))

	for id := 0; id < len(sourceValue); id++ {
		if err := m.addDecoder(sourceValue[id], fmt.Sprintf("[%d]", id)).decode(targetSlice.Index(id)); err != nil {
			return err
		}
	}

	target.Set(targetSlice)
	return nil
}

func (m *Decoder) decodeMap(target reflect.Value) error {
	sourceMap, err := m.Map()
	if err != nil {
		return err
	}

	sourceValue := reflect.Indirect(reflect.ValueOf(sourceMap))

	targetType := target.Type()
	targetKeyType := targetType.Key()
	targetElemType := targetType.Elem()
	targetMap := reflect.MakeMapWithSize(target.Type(), len(sourceMap))

	for _, key := range sourceValue.MapKeys() {
		currentKey := reflect.Indirect(reflect.New(targetKeyType))
		if err := m.addDecoder(key.Interface(), fmt.Sprintf("[%v]", key)).decode(currentKey); err != nil {
			return err
		}

		currentValue := reflect.Indirect(reflect.New(targetElemType))
		if err := m.addDecoder(sourceValue.MapIndex(key).Interface(), fmt.Sprintf("[%v]", key)).decode(currentValue); err != nil {
			return err
		}

		targetMap.SetMapIndex(currentKey, currentValue)
	}

	target.Set(targetMap)
	return nil
}

func (m *Decoder) decodeStruct(target reflect.Value) error {
	ty := target.Type()
	for id := 0; id < ty.NumField(); id++ {
		field := ty.Field(id)
		targetFieldName := field.Name
		if rename, ok := field.Tag.Lookup(m.opts.tags.Rename); ok {
			targetFieldName = rename
		}

		sourceValue, err := m.Get(targetFieldName)
		if err != nil {
			required := false
			if tag, ok := field.Tag.Lookup(m.opts.tags.Required); ok {
				v, err := strconv.ParseBool(tag)
				if err != nil {
					return m.errUnmarshal(err)
				}

				required = v
			}

			if required {
				return err
			}

			if defaultValue, ok := field.Tag.Lookup(m.opts.tags.Default); ok {
				zero := reflect.New(target.Field(id).Type())
				zeroInterface := zero.Interface()
				if err := yaml.Unmarshal([]byte(defaultValue), zeroInterface); err != nil {
					return m.errUnmarshal(err)
				}

				target.Field(id).Set(reflect.ValueOf(zeroInterface).Elem())
				continue
			}
		}

		if err := m.addDecoder(sourceValue, field.Name).decode(target.Field(id)); err != nil {
			return err
		}
	}

	return nil
}

func (m *Decoder) decodePtr(target reflect.Value) error {
	// Create an element of the concrete (non pointer) type and decode
	// into that. Then set the value of the pointer to this type.
	targetType := target.Type()
	targetElemType := targetType.Elem()

	targetRealValue := target
	if targetRealValue.IsNil() {
		targetRealValue = reflect.New(targetElemType)
	}

	if err := m.decode(reflect.Indirect(targetRealValue)); err != nil {
		return err
	}

	target.Set(targetRealValue)
	return nil
}

func (m *Decoder) SliceOf(ty reflect.Type) ([]interface{}, error) {
	if src, ok := (m.v).([]interface{}); ok {
		return src, nil
	}

	return nil, fmt.Errorf("`%s` expected to have type `%s`, but actual type is `%T`", m.path, ty, m.v)
}

func (m *Decoder) Get(key string) (interface{}, error) {
	node, err := m.Map()
	if err != nil {
		return nil, err
	}

	for _, k := range mapKeys(key) {
		if v, ok := node[k]; ok {
			return v, nil
		}
	}

	return nil, fmt.Errorf("`%s` is required", fmt.Sprintf("%s.%s", m.path, key))
}

func mapKeys(v string) []string {
	return []string{
		v,
		strings.ToLower(v),
		toSnakeCase(v),
	}
}

func (m *Decoder) Map() (map[interface{}]interface{}, error) {
	if src, ok := (m.v).(map[interface{}]interface{}); ok {
		return src, nil
	}
	return nil, fmt.Errorf("type assertion to map[interface]interface{} failed, actual value: %T", m.v)
}

func (m *Decoder) Value() reflect.Value {
	return reflect.ValueOf(m.v)
}

func (m *Decoder) ensureType(ty reflect.Type) error {
	if reflect.TypeOf(m.v).Kind() == ty.Kind() {
		return nil
	}

	return m.errTypeMismatch(ty.Name())
}

func (m *Decoder) errUnmarshal(err error) error {
	return fmt.Errorf("`%s`: %v", m.path, err)
}

func (m *Decoder) errTypeMismatch(expectedType string) error {
	return fmt.Errorf("`%s` expected type `%s`, but actual type is `%T`", m.path, expectedType, m.v)
}

func (m *Decoder) errLengthMismatch(expected, actual int) error {
	return fmt.Errorf("`%s` expected length `%d`, but actual length is `%d`", m.path, expected, actual)
}

func (m *Decoder) errOverflow(expectedType string, value interface{}) error {
	return fmt.Errorf("`%s` expected type `%s`, but actual value %v overflows it", m.path, expectedType, value)
}

// Load loads the config from the specified reader into "cfg" variable.
//
// It initially loads content from the reader with following decoding into an
// untyped interface.
//
// Currently it supports the following formats: YAML.
func Load(rd io.Reader, cfg interface{}, options ...Option) error {
	target := reflect.ValueOf(cfg)
	if target.Kind() != reflect.Ptr {
		return fmt.Errorf("`cfg` must be a pointer")
	}

	if !target.Elem().CanAddr() {
		return fmt.Errorf("`cfg` must be addressable, i.e. a pointer")
	}

	data, err := ioutil.ReadAll(rd)
	if err != nil {
		return err
	}

	var v interface{}
	if err := yaml.Unmarshal(data, &v); err != nil {
		return err
	}

	if err := NewDecoder(v, options...).Decode(cfg); err != nil {
		return fmt.Errorf("failed to decode: %v", err)
	}

	if v, ok := cfg.(Validate); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func FromFile(path string, cfg interface{}, options ...Option) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return Load(bytes.NewReader(data), cfg, options...)
}
