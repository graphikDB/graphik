package k8s

import (
	"context"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

type Provider struct {
	namespace string
	clientset *kubernetes.Clientset
}

func NewInClusterProvider(namespace string) (*Provider, error) {
	if namespace == "" {
		namespace = "default"
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get in cluster config")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get in cluster clientset")
	}
	return &Provider{
		namespace: namespace,
		clientset: clientset,
	}, nil
}

func NewOutOfClusterProvider(namespace string) (*Provider, error) {
	if namespace == "" {
		namespace = "default"
	}
	dir, _ := homedir.Dir()
	kubeconfig := filepath.Join(dir, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get in cluster config")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get in cluster clientset")
	}
	return &Provider{
		namespace: namespace,
		clientset: clientset,
	}, nil
}

func (p *Provider) Pods(ctx context.Context) (map[string]string, error) {
	pods, err := p.clientset.CoreV1().Pods(p.namespace).List(
		ctx,
		metav1.ListOptions{
			LabelSelector: "app = graphik",
		})
	if err != nil {
		return nil, err
	}
	data := map[string]string{}
	p.loopPods(pods, data)
	return data, nil
}

func (p *Provider) loopPods(pods *v1.PodList, addrs map[string]string) {
	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning {
			continue
		}
		//for _, condition := range pod.Status.Conditions {
		//	if condition.Type == v1.PodReady && condition.Status != v1.ConditionTrue {
		//		p.loopPods(pods, addrs)
		//	}
		//}
		addr := pod.Status.PodIP
		if addr == "" {
			continue
		}

		addrs[pod.Name] = addr
	}
}
