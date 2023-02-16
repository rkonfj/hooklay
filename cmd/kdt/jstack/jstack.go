package jstack

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rkonfj/opkit/pkg/config"
	"github.com/rkonfj/opkit/pkg/jifa"
	"github.com/rkonfj/opkit/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Cmd = cobra.Command{
	Use:              "jstack",
	Short:            "Dump target environment jvm deployment's stack",
	Example:          "kdt jstack production order",
	Run:              diff,
	TraverseChildren: true,
	Args:             cobra.ExactArgs(2),
}

var enableJifa string

func init() {
	Cmd.Flags().StringVarP(&enableJifa, "jifa", "j", "", "jifa server for upload jstack")
}

func diff(cmd *cobra.Command, args []string) {
	ctx := config.Conf.Environments[args[0]]
	if ctx == nil {
		logrus.Error("can't get environment ", args[0])
		return
	}

	deployName := args[1]

	restConfig := k8s.RestConfig(ctx.To.Context)

	toClientset, err := k8s.Client(ctx.To.Context)
	if err != nil {
		logrus.Fatal(err)
	}

	rpcctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	podsCli := toClientset.CoreV1().Pods(ctx.To.Namespace)

	l, err := podsCli.List(rpcctx, v1.ListOptions{LabelSelector: "app=" + deployName})

	if err != nil {
		logrus.Error(err)
		return
	}
	for _, pod := range l.Items {
		if enableJifa != "" {
			stackBuf := &bytes.Buffer{}
			k8s.Exec(toClientset, restConfig, pod.Name, ctx.To.Namespace, stackBuf, "jstack -l 1")
			err = jifa.Upload(pod.Name, stackBuf, "THREAD_DUMP")
			if err != nil {
				logrus.Error(err)
			}
			continue
		}
		filename := fmt.Sprintf("%d-%s-jstack.txt", time.Now().Unix(), pod.Name)
		fi, err := os.Create(filename)
		if err != nil {
			logrus.Warn(err)
			continue
		}
		k8s.Exec(toClientset, restConfig, pod.Name, ctx.To.Namespace, fi, "jstack -l 1")
	}
}
