package jsonx

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/shippomx/zard/internal/json"
)

// Marshal marshals v into json bytes.
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalToString marshals v into a string.
func MarshalToString(v any) (string, error) {
	data, err := Marshal(v)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Unmarshal unmarshals data bytes into v.
func Unmarshal(data []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := json.UnmarshalUseNumber(decoder, v); err != nil {
		return formatError(string(data), err)
	}

	return nil
}

// UnmarshalFromString unmarshals v from str.
func UnmarshalFromString(str string, v any) error {
	decoder := json.NewDecoder(strings.NewReader(str))
	if err := json.UnmarshalUseNumber(decoder, v); err != nil {
		return formatError(str, err)
	}

	return nil
}

// UnmarshalFromReader unmarshals v from reader.
func UnmarshalFromReader(reader io.Reader, v any) error {
	var buf strings.Builder
	teeReader := io.TeeReader(reader, &buf)
	decoder := json.NewDecoder(teeReader)
	if err := json.UnmarshalUseNumber(decoder, v); err != nil {
		return formatError(buf.String(), err)
	}

	return nil
}

func formatError(v string, err error) error {
	return fmt.Errorf("string: `%s`, error: `%w`", v, err)
}
