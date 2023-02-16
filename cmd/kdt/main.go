package main

import (
	"os"

	"github.com/rkonfj/opkit/cmd/kdt/diff"
	"github.com/rkonfj/opkit/cmd/kdt/envs"
	"github.com/rkonfj/opkit/cmd/kdt/image"
	"github.com/rkonfj/opkit/cmd/kdt/jstack"
	"github.com/rkonfj/opkit/cmd/kdt/jvmgc"
	"github.com/rkonfj/opkit/cmd/kdt/push"
	"github.com/rkonfj/opkit/cmd/kdt/restart"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cobra.Command{
		Use:   "kdt",
		Short: "kubernetes based deployments tools",
	}

	logrus.SetLevel(getLogLevel())
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		DisableQuote:           true,
	})

	rootCmd.AddCommand(&diff.Cmd)
	rootCmd.AddCommand(&push.Cmd)
	rootCmd.AddCommand(&envs.Cmd)
	rootCmd.AddCommand(&image.Cmd)
	rootCmd.AddCommand(&jstack.Cmd)
	rootCmd.AddCommand(&restart.Cmd)
	rootCmd.AddCommand(&jvmgc.Cmd)

	rootCmd.Execute()
}

func getLogLevel() logrus.Level {
	l := os.Getenv("LOG_LEVEL")
	if l == "" {
		l = "info"
	}
	level, err := logrus.ParseLevel(l)
	if err != nil {
		panic(err)
	}
	return level
}
