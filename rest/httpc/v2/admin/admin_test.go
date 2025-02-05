package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestClient struct {
	B bool
}

func TestAddService(t *testing.T) {
	c := &TestClient{}
	RegisterClient(c, func() {
		c.B = true
	})
	AddService(func(client interface{}, fn func()) {
		_ = client.(*TestClient)
		fn()
	})
	assert.True(t, c.B)
}
