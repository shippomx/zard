package docker

import "github.com/shippomx/zard/tools/goctl/internal/cobrax"

var (
	varExeName        string
	varStringGo       string
	varStringBase     string
	varIntPort        int
	varMetricsIntPort int
	varStringHome     string
	varStringRemote   string
	varStringBranch   string
	varStringVersion  string
	varStringTZ       string

	// Cmd describes a docker command.
	Cmd = cobrax.NewCommand("docker", cobrax.WithRunE(dockerCommand))
)

func init() {
	dockerCmdFlags := Cmd.Flags()
	dockerCmdFlags.StringVarWithDefaultValue(&varExeName, "exe", "app")
	dockerCmdFlags.StringVar(&varStringGo, "go")
	dockerCmdFlags.StringVarWithDefaultValue(&varStringBase, "base", "nexus-dev-image.fulltrust.link/base-images/alpine:latest")
	dockerCmdFlags.IntVarWithDefaultValue(&varIntPort, "port", 8080)
	dockerCmdFlags.IntVarWithDefaultValue(&varMetricsIntPort, "metric-port", 8081)
	dockerCmdFlags.StringVar(&varStringHome, "home")
	dockerCmdFlags.StringVar(&varStringRemote, "remote")
	dockerCmdFlags.StringVar(&varStringBranch, "branch")
	dockerCmdFlags.StringVarWithDefaultValue(&varStringVersion, "version", "1.21")
	dockerCmdFlags.StringVar(&varStringTZ, "tz")
}
