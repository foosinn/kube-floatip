package k8s

import (
	"context"
	"fmt"
	"time"
        "os"
        "path/filepath"

	"github.com/foosinn/kube-floatip/internal/config"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

)

type (
	StartLeadingFunc func(context.Context)
	StopLeadingFunc  func()
	NewLeaderFunc    func(string)

	K8sLeaderCallbacks struct {
		OnStartedLeading StartLeadingFunc
		OnStoppedLeading StopLeadingFunc
		OnNewLeader NewLeaderFunc

	}

	K8s struct {
		client   *kubernetes.Clientset
		config   *config.Config
	}
)

func New(config *config.Config) (k *K8s, err error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
          homeDir := os.Getenv("HOME")
          kubeconfig := filepath.Join(homeDir, ".kube", "config")
          cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to load k8s config: %s", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to kubernetes: %s", err)
	}

	k = &K8s{
		client:   client,
		config:   config,
	}
	return
}

func (k *K8s) RunLeaderElection(ctx context.Context, identity string, name string, callbacks K8sLeaderCallbacks) (err error) {
	lock, err := resourcelock.New(
		resourcelock.ConfigMapsResourceLock,
		k.config.Namespace,
		name,
		k.client.CoreV1(),
		k.client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: identity,
		},
	)
	if err != nil {
		return
	}
	lec := leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: callbacks.OnStartedLeading,
			OnStoppedLeading: callbacks.OnStoppedLeading,
			OnNewLeader:      callbacks.OnNewLeader,
		},
	}
	le, err := leaderelection.NewLeaderElector(lec)
	if err != nil {
		return
	}
	if lec.WatchDog != nil {
		lec.WatchDog.SetLeaderElection(le)
	}
	le.Run(ctx)
	return
}

