package stat

import (
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestRemoteWriter(t *testing.T) {
	defer gock.Off()

	gock.New("http://foo.com").Reply(200).BodyString("foo")
	writer := NewRemoteWriter("http://foo.com")
	err := writer.Write(&StatReport{
		Name: "bar",
	})
	assert.Nil(t, err)
}

func TestRemoteWriterFail(t *testing.T) {
	defer gock.Off()

	gock.New("http://foo.com").Reply(503).BodyString("foo")
	writer := NewRemoteWriter("http://foo.com")
	err := writer.Write(&StatReport{
		Name: "bar",
	})
	assert.NotNil(t, err)
}
