package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/foosinn/kube-floatip/internal/config"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/workqueue"
)

type (
	StartLeadingFunc func(context.Context)
	StopLeadingFunc  func()
	NewLeaderFunc    func(string)

	K8s struct {
		client   *kubernetes.Clientset
		config   *config.Config
		queue    workqueue.RateLimitingInterface
		lister   listerv1.NodeLister
		informer cache.Controller
	}
)

func New(config *config.Config) (k *K8s, err error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load inclusterconfig: %s", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to kubernetes: %s", err)
	}

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	indexer, informer := cache.NewIndexerInformer(
		cache.NewListWatchFromClient(client.RESTClient(), "nodes", v1.NamespaceAll, fields.Everything()),
		&v1.Node{},
		10*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				queueAppend(obj, queue)
			},
			UpdateFunc: func(old, new interface{}) {
				queueAppend(new, queue)
			},
			DeleteFunc: func(obj interface{}) {
				queueAppend(obj, queue)
			},
		},
		cache.Indexers{},
	)
	lister := listerv1.NewNodeLister(indexer)

	k = &K8s{
		client:   client,
		config:   config,
		queue:    queue,
		lister:   lister,
		informer: informer,
	}
	return
}

func queueAppend (obj interface{}, queue workqueue.RateLimitingInterface) {
	if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
		queue.Add(key)
	}
}

func (k *K8s) RunLeaderElection(ctx context.Context, leading StartLeadingFunc, stopping StopLeadingFunc, new NewLeaderFunc) (err error) {
	lock, err := resourcelock.New(
		resourcelock.ConfigMapsResourceLock,
		k.config.Namespace,
		k.config.Name,
		k.client.CoreV1(),
		k.client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: k.config.Id,
		},
	)
	lec := leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: leading,
			OnStoppedLeading: stopping,
			OnNewLeader:      new,
		},
	}
	leaderelection.RunOrDie(ctx, lec)
	return
}

func (k *K8s) GetNodeAnnotation(node string, annotation string) (value string, err error) {
	n, err := k.lister.Get(node)
	if err != nil {
		return "", fmt.Errorf("unable to get annotation: %s", err)
	}
	value, ok := n.GetAnnotations()[annotation]
	if !ok {
		return "", fmt.Errorf("annotation not found")
	}
	k.client.CoreV1().Nodes().Get(k.config.Id, metav1.GetOptions{})
        return
}
