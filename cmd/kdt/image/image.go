package image

import (
	"context"
	"fmt"
	"time"

	"github.com/rkonfj/opkit/pkg/config"
	"github.com/rkonfj/opkit/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Cmd = cobra.Command{
	Use:     "image",
	Short:   "Show environment's images",
	Example: "kdt image production",
	Run:     run,
	Args:    cobra.ExactArgs(1),
}

func run(cmd *cobra.Command, args []string) {
	ctx := config.Conf.Environments[args[0]]
	if ctx == nil {
		logrus.Error("can't get environment ", args[0])
		return
	}

	fromClientset, err := k8s.Client(ctx.From.Context)
	if err != nil {
		logrus.Fatal(err)
	}

	clisentset, err := k8s.Client(ctx.To.Context)
	if err != nil {
		logrus.Fatal(err)
	}

	rpcctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fromCli := fromClientset.AppsV1().Deployments(ctx.From.Namespace)
	cli := clisentset.AppsV1().Deployments(ctx.To.Namespace)

	l, err := cli.List(rpcctx, v1.ListOptions{})
	if err != nil {
		logrus.Error(err)
		return
	}

	for _, item := range l.Items {
		deploy, err := fromCli.Get(rpcctx, item.Name, v1.GetOptions{})
		if err != nil {
			if !config.Conf.IgnoreWarnning {
				logrus.Warn(err)
			}
			continue
		}
		var fromContainers map[string]string = make(map[string]string)
		for _, container := range deploy.Spec.Template.Spec.Containers {
			fromContainers[container.Name] = container.Image
		}
		for _, container := range item.Spec.Template.Spec.Containers {
			diff := ""
			if container.Image != fromContainers[container.Name] {
				diff = "⬆"
			}
			fmt.Printf("%s:%s\n    ➡ %s %s\n", container.Name, container.Image, fromContainers[container.Name], diff)
		}
	}
}
