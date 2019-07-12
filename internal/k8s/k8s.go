package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/foosinn/kube-floatip/internal/config"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

)

type (
	StartLeadingFunc func(context.Context)
	StopLeadingFunc  func()
	NewLeaderFunc    func(string)

	K8s struct {
		client   *kubernetes.Clientset
		config   *config.Config
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

	k = &K8s{
		client:   client,
		config:   config,
	}
	return
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

