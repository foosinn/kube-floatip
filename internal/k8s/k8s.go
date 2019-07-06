package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/foosinn/kube-floatip/internal/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/transport"
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
	cfg.Wrap(transport.ContextCanceller(ctx, fmt.Errorf("the leader is shutting down")))
	client := kubernetes.NewForConfigOrDie(cfg)

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name: config.Name,
			Namespace: config.Namespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: config.Id,
		},
	}
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

	_, err = client.CoordinationV1().Leases(config.Namespace).Get(config.Name, metav1.GetOptions{})
	if err == nil || !strings.Contains(err.Error(), "the leader is shutting down") {
		return fmt.Errorf("expected an error when checking lease: %+v", err)
	}
	return
}
