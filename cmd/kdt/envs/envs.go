package envs

import (
	"fmt"

	"github.com/rkonfj/opkit/pkg/config"
	"github.com/spf13/cobra"
)

var Cmd = cobra.Command{
	Use:     "envs",
	Short:   "Show all environments",
	Example: "kdt envs",
	Run:     run,
}

func run(cmd *cobra.Command, args []string) {
	for k, v := range config.Conf.Environments {
		fmt.Printf("%s (%s:%s > %s:%s)\n", k, v.From.Context, v.From.Namespace, v.To.Context, v.To.Namespace)
	}
}
