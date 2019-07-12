package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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

	k, err := k8s.New(cfg)
	if err != nil {
		klog.Fatal(err)
	}
	k.RunLeaderElection(
		ctx,
		func(ctx context.Context) {
			klog.Infof("%s: I am the leader", cfg.Id)
			<-ctx.Done()
			klog.Infof("%s: No more leader", cfg.Id)

		},
		func() {
			klog.Infof("%s: stopping", cfg.Id)
		},
		func(i string) {
			klog.Infof("%s: new leader is %s", cfg.Id, i)
		},
	)
}
