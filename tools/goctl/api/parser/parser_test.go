package parser

import (
	_ "embed"
	"testing"

	"github.com/shippomx/zard/tools/goctl/api/spec"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/test.api
var testApi string

func TestParseContent(t *testing.T) {
	sp, err := ParseContent(testApi)
	assert.Nil(t, err)
	assert.Equal(t, spec.Doc{`// syntax doc`}, sp.Syntax.Doc)
	assert.Equal(t, spec.Doc{`// syntax comment`}, sp.Syntax.Comment)
	for _, tp := range sp.Types {
		if tp.Name() == "Request" {
			assert.Equal(t, []string{`// type doc`}, tp.Documents())
		}
	}
	for _, e := range sp.Service.Routes() {
		if e.Handler == "GreetHandler" {
			assert.Equal(t, spec.Doc{"// handler doc"}, e.HandlerDoc)
			assert.Equal(t, spec.Doc{"// handler comment"}, e.HandlerComment)
		}
	}
}

func TestMissingService(t *testing.T) {
	sp, err := ParseContent("")
	assert.Nil(t, err)
	err = sp.Validate()
	assert.Equal(t, spec.ErrMissingService, err)
}

func TestValidateAliasType(t *testing.T) {
	tests := []struct {
		name         string
		expected     []string
		expectedFlag bool
	}{
		{"int", []string{}, true},
		{"any", []string{}, true},
		{"interface{}", []string{}, true},
		{"[]int", []string{}, true},
		{"*int", []string{}, true},
		{"map[string]int", []string{}, true},
		{"map[string]map[int]int", []string{}, true},
		{"invalid", []string{"invalid"}, false},
		{"map[", []string{"map["}, false},
		{"map[]", []string{}, false},
	}

	for _, tt := range tests {
		actual, actualFlag := ValidateAliasType(tt.name)
		if !compareSlices(actual, tt.expected) || actualFlag != tt.expectedFlag {
			t.Errorf("ValidateAliasType(%q) = (%v, %v), want (%v, %v)", tt.name, actual, actualFlag, tt.expected, tt.expectedFlag)
		}
	}
}

func compareSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
