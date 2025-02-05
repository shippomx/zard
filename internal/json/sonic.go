//go:build (linux || windows || darwin) && amd64 && !stdjson
// +build linux windows darwin
// +build amd64
// +build !stdjson

package json

import (
	stdjson "encoding/json"

	"github.com/bytedance/sonic"
)

// Name is the name of the effective json package.
const Name = "sonic"

var (
	json          = sonic.ConfigStd
	Marshal       = json.Marshal
	Unmarshal     = json.Unmarshal
	MarshalIndent = json.MarshalIndent
	NewDecoder    = json.NewDecoder
	NewEncoder    = json.NewEncoder
)

type Number = stdjson.Number
type Decoder = sonic.Decoder

func UnmarshalUseNumber(decoder Decoder, v any) error {
	decoder.UseNumber()
	return decoder.Decode(v)
}
