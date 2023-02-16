package push

import (
	"bytes"
	"context"
	"fmt"
	"log"
	syn "sync"
	"time"

	"github.com/rkonfj/opkit/pkg/config"
	"github.com/rkonfj/opkit/pkg/gitlab"
	"github.com/rkonfj/opkit/pkg/internalutil"
	"github.com/rkonfj/opkit/pkg/k8s"
	"github.com/rkonfj/opkit/pkg/notification"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

var Cmd = cobra.Command{
	Use:     "push",
	Short:   "Push different deployments from A to B",
	Example: "kdt push production [order] [pay]",
	Run:     sync,
	Args:    cobra.MinimumNArgs(1),
}

var (
	wait        bool
	gitlabGroup map[string]int
)

func init() {
	Cmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for deploy update successfully")
}

func sync(cmd *cobra.Command, args []string) {
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

	var oldDeploys map[string]string = make(map[string]string)
	var newDeploys map[string]string = make(map[string]string)

	if len(args) > 1 {
		for _, mod := range args[1:] {
			fromDeploy, err := fromCli.Get(rpcctx, mod, v1.GetOptions{})
			if err != nil {
				logrus.Warn(err)
				continue
			}
			toDeploy, err := toCli.Get(rpcctx, mod, v1.GetOptions{})
			if err != nil {
				logrus.Warn(err)
				continue
			}
			oldImage, newImage := applyDeployment(*fromDeploy, *toDeploy, toCli)
			if oldImage != "" {
				oldDeploys[mod] = oldImage
				newDeploys[mod] = newImage
			}
		}
		if wait {
			waitDeploysSuccess(oldDeploys, newDeploys, toCli, ctx)
		}
		return
	}

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
		oldImage, newImage := applyDeployment(*deploy, item, toCli)
		if oldImage != "" {
			oldDeploys[item.Name] = oldImage
			newDeploys[item.Name] = newImage
		}
	}
	if wait {
		waitDeploysSuccess(oldDeploys, newDeploys, toCli, ctx)
	}
}

func applyDeployment(fromDeploy, toDeploy appsv1.Deployment, toCli typedappsv1.DeploymentInterface) (string, string) {
	rpcctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var fromContainers map[string]string = make(map[string]string)

	for _, container := range fromDeploy.Spec.Template.Spec.Containers {
		fromContainers[container.Name] = container.Image
	}

	var toContainers map[string]string = make(map[string]string)

	newImage := ""

	for _, container := range toDeploy.Spec.Template.Spec.Containers {
		if container.Image != fromContainers[container.Name] {
			toContainers[container.Name] = container.Image
			for idx, newContainer := range fromDeploy.Spec.Template.Spec.Containers {
				if newContainer.Name == container.Name {
					toDeploy.Spec.Template.Spec.Containers[idx].Image = fromContainers[container.Name]
				}
			}
			_, err := toCli.Update(rpcctx, &toDeploy, v1.UpdateOptions{})
			if err != nil {
				logrus.Error(err)
				continue
			}
			log.Println(container.Name, "updated to", fromContainers[container.Name])
			newImage = fromContainers[container.Name]
		}
	}
	if len(toContainers) > 0 {
		return toContainers[fromDeploy.Name], fromContainers[fromDeploy.Name]
	}
	if newImage != "" {
		return toDeploy.Spec.Template.Spec.Containers[0].Image, newImage
	}
	return "", ""
}

func waitDeploysSuccess(oldDeploys, newDeploys map[string]string, toCli typedappsv1.DeploymentInterface, ctx *config.DeployContext) {
	fmt.Println("Waiting deployments ready ...")

	var changelog bytes.Buffer
	var rw syn.RWMutex
	wg := syn.WaitGroup{}
	wg.Add(len(newDeploys))
	for k, v := range newDeploys {
		go func(deployName, oldImage, newImage string) {
			for {
				rpcctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				time.Sleep(5 * time.Second)
				deploy, err := toCli.Get(rpcctx, deployName, v1.GetOptions{})
				if err != nil {
					logrus.Error(err)
					return
				}
				if deploy.Status.ReadyReplicas == deploy.Status.Replicas && deploy.Status.UnavailableReplicas == 0 {
					break
				}
			}
			wg.Done()
			logrus.Info(deployName, " is ready")
			projectId, exists := gitlabGroup[deployName]
			if !exists {
				logrus.Error("no project mapping for ", deployName)
				return
			}
			cl := gitlab.Changelog(internalutil.ParseCommitIDFromImage(oldImage), internalutil.ParseCommitIDFromImage(newImage), projectId)
			logrus.Debug("oldImage ", oldImage, "; newImage ", newImage, "; projectId ", projectId, "; changelog ", cl)
			rw.Lock()
			changelog.WriteString(fmt.Sprintf("### ** %s **\n", deployName))
			changelog.WriteString(cl)
			changelog.WriteString("\n---")
			rw.Unlock()
		}(k, oldDeploys[k], v)
	}
	wg.Wait()
	if len(newDeploys) == 0 {
		return
	}
	err := notification.SendDingtalk(fmt.Sprintf("%s:%s > %s:%s", ctx.From.Context, ctx.From.Namespace, ctx.To.Context, ctx.To.Namespace), changelog.String())
	if err != nil {
		logrus.Error(err)
	}
}
