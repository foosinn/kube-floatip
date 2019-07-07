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
	StopLeadingFunc func()
	NewLeaderFunc func(string)

)

func RunLeaderElection(ctx context.Context, config *config.Config, leading StartLeadingFunc,
	stopping StopLeadingFunc, new NewLeaderFunc) (err error) {

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to load inclusterconfig: %s", err)
	}
	client := kubernetes.NewForConfigOrDie(cfg)

	lock, err := resourcelock.New(
		resourcelock.ConfigMapsResourceLock,
		config.Namespace,
		config.Name,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: config.Id,
		},
	)
	lec := leaderelection.LeaderElectionConfig{
		Lock: lock,
		ReleaseOnCancel: true,
		LeaseDuration: 60 * time.Second,
		RenewDeadline: 15 * time.Second,
		RetryPeriod: 5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: leading,
			OnStoppedLeading: stopping,
			OnNewLeader: new,
		},
	}
	leaderelection.RunOrDie(ctx, lec)
	return
}
