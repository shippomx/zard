package lint

import (
	"github.com/spf13/cobra"
	"testing"
)

func Test_lint(t *testing.T) {
	type args struct {
		in0 *cobra.Command
		in1 []string
	}
	VarStringDir = "/Users/zippy/Documents/goctl_test"
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test_lint",
			args: args{
				in0: nil,
				in1: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := lint(tt.args.in0, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("lint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
