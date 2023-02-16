package k8s

import (
	"io"
	"os"

	"github.com/rkonfj/opkit/pkg/config"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

func Client(context string) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(RestConfig(context))
}

func RestConfig(context string) *rest.Config {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: config.Conf.Kubeconfig},
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	return config
}

func Exec(client *kubernetes.Clientset, config *rest.Config, podName string, namespace string, out io.Writer, cmd string) {
	req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(namespace).SubResource("exec")
	option := &corev1.PodExecOptions{
		Command: []string{
			"sh",
			"-c",
			cmd,
		},
		Stdin:  false,
		Stdout: true,
		Stderr: true,
		TTY:    true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		logrus.Fatal(err)
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: out,
		Stderr: os.Stderr,
	})
	if err != nil {
		logrus.Fatal(err)
	}
}
