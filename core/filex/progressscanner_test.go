package filex

import (
	"strings"
	"testing"

	"github.com/cheggaaa/pb/v3"
	"github.com/stretchr/testify/assert"
)

func TestProgressScanner(t *testing.T) {
	const text = "hello, world"
	bar := pb.New(100)
	var builder strings.Builder
	builder.WriteString(text)
	scanner := NewProgressScanner(&mockedScanner{builder: &builder}, bar)
	assert.True(t, scanner.Scan())
	assert.Equal(t, text, scanner.Text())
}

type mockedScanner struct {
	builder *strings.Builder
}

func (s *mockedScanner) Scan() bool {
	return s.builder.Len() > 0
}

func (s *mockedScanner) Text() string {
	return s.builder.String()
}
