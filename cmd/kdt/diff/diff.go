package diff

import (
	"context"
	"fmt"
	"time"

	"github.com/rkonfj/opkit/pkg/config"
	"github.com/rkonfj/opkit/pkg/gitlab"
	"github.com/rkonfj/opkit/pkg/internalutil"
	"github.com/rkonfj/opkit/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Cmd = cobra.Command{
	Use:     "diff",
	Short:   "Show different deployments from A to B",
	Example: "kdt diff production",
	Run:     diff,
	Args:    cobra.ExactArgs(1),
}

var (
	showCommits bool
	gitlabGroup map[string]int
)

func init() {
	Cmd.Flags().BoolVarP(&showCommits, "showCommits", "s", false, "show diff commits from gitlab")
}

func diff(cmd *cobra.Command, args []string) {
	ctx := config.Conf.Environments[args[0]]
	if ctx == nil {
		logrus.Error("can't get environment ", args[0])
		return
	}

	if g, exists := config.Conf.Gitlab.Groups[ctx.GitlabGroup]; !exists {
		logrus.Error("can't get gitlab group: ", ctx.GitlabGroup)
		return
	} else {
		gitlabGroup = g
	}

	fromClientset, err := k8s.Client(ctx.From.Context)
	if err != nil {
		logrus.Fatal(err)
	}

	toClientset, err := k8s.Client(ctx.To.Context)
	if err != nil {
		logrus.Fatal(err)
	}

	rpcctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fromCli := fromClientset.AppsV1().Deployments(ctx.From.Namespace)
	toCli := toClientset.AppsV1().Deployments(ctx.To.Namespace)

	l, err := toCli.List(rpcctx, v1.ListOptions{})
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
			if container.Image != fromContainers[container.Name] {
				fmt.Println("🚀", container.Image, ">", fromContainers[container.Name])
				if showCommits {
					showCommitsFromGitlab(container.Name, container.Image, fromContainers[container.Name])
				}
			}
		}
	}
}

func showCommitsFromGitlab(name, oldContainerImage, newContainerImage string) {
	projectId, exists := gitlabGroup[name]
	if !exists {
		logrus.Error("no project mapping for ", name)
		return
	}
	fmt.Println(gitlab.Changelog(internalutil.ParseCommitIDFromImage(oldContainerImage), internalutil.ParseCommitIDFromImage(newContainerImage), projectId))
}
