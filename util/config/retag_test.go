package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToSnakeCase(t *testing.T) {
	batch := map[string]string{
		"":                           "",
		"t":                          "t",
		"test":                       "test",
		"testC":                      "test_c",
		"testCase":                   "test_case",
		"TestCase":                   "test_case",
		"Test Case":                  "test_case",
		" TestCase\t":                "test_case",
		" Test Case":                 "test_case",
		"Test Case ":                 "test_case",
		" Test Case ":                "test_case",
		"test_case":                  "test_case",
		"Test":                       "test",
		"HTTPStatusCode":             "http_status_code",
		"ParseURL.DoParse":           "parse_url.do_parse",
		"Convert Space":              "convert_space",
		"Convert-dash":               "convert_dash",
		"Skip___MultipleUnderscores": "skip_multiple_underscores",
		"Skip   MultipleSpaces":      "skip_multiple_spaces",
		"Skip---MultipleDashes":      "skip_multiple_dashes",
		"ManyManyWords":              "many_many_words",
		"manyManyWords":              "many_many_words",
		"AnyKind of_string":          "any_kind_of_string",
		"JSONData":                   "json_data",
		"userID":                     "user_id",
	}

	for s, expected := range batch {
		assert.Equal(t, expected, toSnakeCase(s))
	}
}
