package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/foosinn/kube-floatip/internal/config"
	"github.com/foosinn/kube-floatip/internal/floatip"
	"github.com/foosinn/kube-floatip/internal/k8s"

	"k8s.io/klog"
)

func main() {
	klog.SetOutput(os.Stdout)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Create()
	if err != nil {
		klog.Errorf("unable to load config: %s", err)
		cancel()
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-sigs
		klog.Errorf("%s: received signal %s, shutting down", cfg.NodeName, s.String())
		cancel()
	}()

	k, err := k8s.New(cfg)
	if err != nil {
		klog.Error(err)
		cancel()
	}
	if len(cfg.ProviderIPs) == 0 {
		klog.Errorf("No IPs to provide.")
		cancel()
	}
	for _, ip := range cfg.ProviderIPs {
		go provide(ctx, k, cfg, ip)
	}

	<-ctx.Done()
}

func provide(ctx context.Context, k *k8s.K8s, cfg *config.Config, ipID int) (err error) {
	fip, err := floatip.New(ipID, cfg)
	if err != nil {
		klog.Fatal(err)
	}

	var leaderCtx context.Context
	var leaderCancel context.CancelFunc
	err = k.RunLeaderElection(
		ctx,
                cfg.NodeName,
                fip.DnsName(),
		k8s.K8sLeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				leaderCtx, leaderCancel = context.WithCancel(ctx)
				defer leaderCancel()
				klog.Infof("%s: i am leader for '%s'", cfg.NodeName, fip.String())
				err := fip.Bind(leaderCtx)
				if err != nil {
					klog.Error(err)
					klog.Fatalf("%s: unable to bind ip, killing myself", cfg.NodeName)
				}
				klog.Infof("%s: no more leader", cfg.NodeName)
			},
			OnStoppedLeading: func() {
				if leaderCancel == nil {
					klog.Fatalf("%s: cancel function is not set, stopping", cfg.NodeName)
				}
				klog.Infof("%s: removing ips", cfg.NodeName)
				leaderCancel()
			},
			OnNewLeader: func(i string) {
				klog.Infof("%s: new leader for %s is %s", cfg.NodeName, fip.String(), i)
			},
		},
	)
	return err
}
