package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/foosinn/kube-floatip/internal/config"
	"github.com/foosinn/kube-floatip/internal/k8s"

	"k8s.io/klog"
)

func main() {
	klog.SetOutput(os.Stdout)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Create()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-sigs
		klog.Errorf("%s: received signal %s, shutting down", cfg.Id, s.String())
		cancel()
	}()

	k8s.RunLeaderElection(
		ctx,
		cfg,
		func(ctx context.Context) {
			for {
				select {
				case <-time.After(time.Second):
					klog.Infof("%s: I am still the leader", cfg.Id)
				case <-ctx.Done():
					klog.Infof("s%: No more leader", cfg.Id)
					break
				}
			}

		},
		func() {
			klog.Infof("%s: stopping")
		},
		func(string) {},
	)
}
