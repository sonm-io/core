package retag

import (
	"fmt"
	"reflect"
	"strings"
)

// NewView creates TagMaker which makes tag for structure's field by the next rules:
//   - If field's tag contains value with key 'view' and the value matches with
//     value passed in the 'name' parameter, the maker returns the key passed in the 'tag' parameter
//     with its value (if presented) from field's tag or empty string;
//   - If field's tag contains value with key 'view' and the value doesn't match with
//     value passed in the 'name' parameter, the maker returns `<tag>:"-"`;
//   - If field's tag doesn't contain 'view' section, the maker returns value of the original tag.
//
// Section view may contain comma-separated list of views or '*'. '*' matches any view.
//
// Examples for NewView("json", "admin"):
//   ``                  -> ``
//   `xml:"name"`        -> `xml:"name"`
//   `view:"-"`          -> `json:"-"`
//   `view:"user"`       -> `json:"-"`
//   `view:"*"`          -> ``
//   `view:"admin"`      -> ``
//   `view:"user,admin"` -> ``
//   `view:"admin" json:"Name,omitempty"` -> `json:"Name,omitempty"`
//
// See package examples additionally.
func NewView(tag, name string) TagMaker {
	return tagView{name, tag}
}

type tagView struct {
	name string
	tag  string
}

func (v tagView) MakeTag(t reflect.Type, fieldIndex int) reflect.StructTag {
	const key = "view"
	field := t.Field(fieldIndex)
	value := field.Tag.Get(key)
	if value == "" {
		return field.Tag
	}
	if v.isMatch(value) {
		defaultValue := field.Tag.Get(v.tag)
		if defaultValue != "" {
			defaultValue = fmt.Sprintf(`%s:"%s"`, v.tag, defaultValue)
		}
		return reflect.StructTag(defaultValue)
	}
	return reflect.StructTag(v.tag + `:"-"`)
}

func (v tagView) isMatch(tag string) bool {
	if tag == "*" {
		return true
	}
	list := parseStringList(tag)
	if list.contains(v.name) {
		return true
	}
	return false
}

func parseStringList(list string) stringList {
	return strings.Split(list, ",")
}

type stringList []string

func (l stringList) contains(s string) bool {
	for _, i := range l {
		if i == s {
			return true
		}
	}
	return false
}
