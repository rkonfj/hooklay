package restart

import (
	"context"
	"fmt"
	"time"

	"github.com/rkonfj/opkit/pkg/config"
	"github.com/rkonfj/opkit/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

var Cmd = cobra.Command{
	Use:              "restart",
	Short:            "Restart target deployments",
	Example:          "kdt restart production [order] [pay]",
	Run:              diff,
	TraverseChildren: true,
	Args:             cobra.MinimumNArgs(1),
}

func diff(cmd *cobra.Command, args []string) {
	ctx := config.Conf.Environments[args[0]]
	if ctx == nil {
		logrus.Error("can't get environment ", args[0])
		return
	}

	toClientset, err := k8s.Client(ctx.To.Context)
	if err != nil {
		logrus.Fatal(err)
	}

	rpcctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	toCli := toClientset.AppsV1().Deployments(ctx.To.Namespace)

	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format(time.RFC3339))

	specialDeploys := args[1:]
	if len(specialDeploys) > 0 {
		for _, deployName := range specialDeploys {
			_, err = toCli.Patch(rpcctx, deployName, k8stypes.StrategicMergePatchType, []byte(data), v1.PatchOptions{})
			if err != nil {
				logrus.Warn(err)
				continue
			}
			fmt.Println(deployName, "restart signal send")
		}
		return
	}

	l, err := toCli.List(rpcctx, v1.ListOptions{})
	if err != nil {
		logrus.Error(err)
		return
	}
	for _, deployment := range l.Items {
		_, err = toCli.Patch(rpcctx, deployment.Name, k8stypes.StrategicMergePatchType, []byte(data), v1.PatchOptions{})
		if err != nil {
			logrus.Error(err)
			continue
		}
		fmt.Println(deployment.Name, "restart signal send")
	}
}
