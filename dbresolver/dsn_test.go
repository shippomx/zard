package dbresolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDSN(t *testing.T) {
	cases := []struct {
		Name        string
		DSN         DSN
		WantTimeout string
	}{
		{
			Name: "check default timeout",
			DSN: DSN{
				Path:     "localhost",
				Port:     3306,
				Dbname:   "mydb",
				Username: "user",
				Password: "pswd",
				Config:   "",
			},
			WantTimeout: "timeout=3s",
		},
		{
			Name: "customize timeout",
			DSN: DSN{
				Path:     "localhost",
				Port:     3306,
				Dbname:   "mydb",
				Username: "user",
				Password: "pswd",
				Config:   "charset=utf8&timeout=10s",
			},
			WantTimeout: "timeout=10s",
		},
	}

	for _, tc := range cases {
		assert.Contains(t, tc.DSN.Str(), tc.WantTimeout)
	}
}
