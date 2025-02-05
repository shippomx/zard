package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shippomx/zard/tools/goctl/util/pathx"
	"github.com/stretchr/testify/assert"
)

func TestProtoTmpl(t *testing.T) {
	_ = Clean()
	// exists dir
	err := ProtoTmpl(pathx.MustTempDir(), false, false)
	assert.Nil(t, err)

	// not exist dir
	dir := filepath.Join(pathx.MustTempDir(), "test")
	err = ProtoTmpl(dir, false, false)
	assert.Nil(t, err)
}
func TestProtoTmplWithError(t *testing.T) {
	err := ProtoTmpl("output.proto", true, false)
	defer pathx.RemoveIfExist("output.proto")
	assert.Nil(t, err)
	//read output.proto
	data, err := os.ReadFile("output.proto")
	assert.Nil(t, err)
	assert.Contains(t, string(data), "validate/validate.proto")

	err = ProtoTmpl("output2.proto", false, false)
	defer pathx.RemoveIfExist("output2.proto")
	assert.Nil(t, err)
	data, err = os.ReadFile("output2.proto")
	assert.Nil(t, err)
	assert.NotContains(t, string(data), "validate/validate.proto")
}

func TestCopyThirdPartyFromEmbed(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := pathx.MustTempDir()
	defer os.RemoveAll(tmpDir)

	// Call the function to copy the files
	err := CopyThirdPartyFromEmbed(filepath.Join(tmpDir, "test.proto"))
	assert.Nil(t, err)
	// check double copy
	err = CopyThirdPartyFromEmbed(filepath.Join(tmpDir, "test.proto"))
	assert.Nil(t, err)
	//check exist third_party dir
	_, err = os.Stat(filepath.Join(tmpDir, "third_party"))
	assert.Nil(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "third_party", "validate", "validate.proto"))
	assert.Nil(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "third_party", "google", "protobuf", "descriptor.proto"))
	assert.Nil(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "third_party", "google", "api", "annotations.proto"))
	assert.Nil(t, err)
}
