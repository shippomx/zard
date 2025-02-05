package client

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListDownload(t *testing.T) {
	token := os.Getenv("DOWNLOAD_TOKEN")
	if token == "" {
		t.SkipNow()
	}
	client := NewClient("gatebackend", "i18n", token)
	res, err := client.ListDownload()
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestCreateDownload(t *testing.T) {
	token := os.Getenv("DOWNLOAD_TOKEN")

	if token == "" {
		t.SkipNow()
	}
	if os.Getenv("DEV") != "true" {
		t.SkipNow()
	}
	file, err := os.CreateTemp("", "test")
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	client := NewClient("gatebackend", "i18n", token)
	err = client.CreateDownloads(file.Name())
	assert.NoError(t, err)
}
