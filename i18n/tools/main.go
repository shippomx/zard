package main

import (
	"github.com/shippomx/zard/i18n/tools/checker"
	"github.com/shippomx/zard/i18n/tools/pusher"
	"github.com/shippomx/zard/i18n/tools/release"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "i18n-tools",
		Short: "This is a app for i18n tools.",
		Long:  ``,
	}

	rootCmd.AddCommand(checker.Init())
	rootCmd.AddCommand(pusher.Init())
	rootCmd.AddCommand(release.Init())
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
	if checker.Inited {
		err := checker.Check()
		if err != nil {
			panic(err)
		}
	}
	if pusher.IpAddr != "" {
		err := pusher.Push()
		if err != nil {
			panic(err)
		}
	}
}
