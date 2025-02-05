package lint

import (
	"github.com/shippomx/zard/tools/goctl/internal/cobrax"
)

func init() {
	var cmdFlags = Cmd.Flags()

	cmdFlags.StringVar(&VarStringDir, "dir")
}

var Cmd = cobrax.NewCommand("lint", cobrax.WithRunE(lint))
