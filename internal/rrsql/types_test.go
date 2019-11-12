package rrsql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNullableIntRedisScan(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    interface{}
		expected NullInt
		errormsg string
	}{
		{"EmptyValue", nil, NullInt{}, ""},
		{"NormalScan", []byte(`{3345678 true}`), NullInt{Int: 3345678, Valid: true}, ""},
		{"InvalidScan", []byte(`{3345678 false}`), NullInt{Int: 0, Valid: false}, ""},
		{"AbsentCurlyBracket", []byte(`3345678 true`), NullInt{Int: 0, Valid: false}, ""},
		{"AssertByteArrayError", `{3345678 true}`, NullInt{}, "RedisScan error assert byte array"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var dest = &NullInt{}
			if err := dest.RedisScan(tc.input); err != nil {
				assert.Equal(t, tc.errormsg, err.Error())
			}
			assert.Equal(t, tc.expected, *dest)
		})
	}
}

func TestNullableStringRedisScan(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    interface{}
		expected NullString
		errormsg string
	}{
		{"EmptySrc", nil, NullString{}, ""},
		{"NormalScan", []byte(`{foo true}`), NullString{String: "foo", Valid: true}, ""},
		{"InvalidScan", []byte(`{ false}`), NullString{String: "", Valid: false}, ""},
		{"AbsentCurlyBracket", []byte(`foo true`), NullString{String: "", Valid: false}, ""},
		{"AssertByteArrayError", `{foo true}`, NullString{String: "", Valid: false}, "RedisScan error assert byte array"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var dest = &NullString{}
			if err := dest.RedisScan(tc.input); err != nil {
				assert.Contains(t, tc.errormsg, err.Error())
			}
			assert.Equal(t, tc.expected, *dest)
		})
	}
}

func TestNullableTimeRedisScan(t *testing.T) {

	for _, tc := range []struct {
		name     string
		input    interface{}
		expected NullTime
		errormsg string
	}{
		{"EmptySrc", nil, NullTime{}, ""},
		{"NormalScan", []byte(`{2018-07-19 02:36:47 +0000 UTC true}`), NullTime{
			Time: time.Date(2018, time.July, 19, 2, 36, 47, 0, time.UTC), Valid: true}, ""},
		{"InvalidScan", []byte(`{2018-07-19 02:36:47 +0000 UTC false}`), NullTime{Time: time.Time{}, Valid: false}, ""},
		{"AbsentCurlyBracket", []byte(`2018-07-19 02:36:47 +0000 UTC true`), NullTime{
			Time: time.Time{}, Valid: false}, ""},
		{"AssertByteArrayError", `{2018-07-19 02:36:47 +0000 UTC true}`, NullTime{}, "RedisScan error assert byte array"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var dest = &NullTime{}
			if err := dest.RedisScan(tc.input); err != nil {
				// Compare error string
				assert.Equal(t, tc.errormsg, err.Error())
			}
			assert.Equal(t, tc.expected, *dest)
		})
	}
}
