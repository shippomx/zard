//go:build stdjson || !(amd64 && (linux || windows || darwin))
// +build stdjson !amd64 !linux,!windows,!darwin

package json

import "encoding/json"

// Name is the name of the effective json package.
const Name = "encoding/json"

var (
	Marshal       = json.Marshal
	Unmarshal     = json.Unmarshal
	MarshalIndent = json.MarshalIndent
	NewDecoder    = json.NewDecoder
	NewEncoder    = json.NewEncoder
)

type Number = json.Number
type Decoder = json.Decoder

func UnmarshalUseNumber(decoder *Decoder, v any) error {
	decoder.UseNumber()
	return decoder.Decode(v)
}
