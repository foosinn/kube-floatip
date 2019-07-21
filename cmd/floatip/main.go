package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"net"
	"fmt"

	"github.com/foosinn/kube-floatip/internal/config"
	"github.com/foosinn/kube-floatip/internal/floatip"
	"github.com/foosinn/kube-floatip/internal/ip"
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

	ipcidr := fmt.Sprintf("%s/32", fip.String())
	if net.ParseIP(fip.String()).To4() == nil {
		ipcidr = fmt.Sprintf("%s/128", fip.String())
	}
	linkIp, err := ip.NewIP(ipcidr, cfg.Link)

	err = k.RunLeaderElection(
		ctx,
                cfg.NodeName,
                fip.DnsName(),
		k8s.K8sLeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				klog.Infof("%s: i am leader for '%s'", cfg.NodeName, fip.String())
				err := linkIp.Bind()
				if err != nil {
					klog.Error(err)
					klog.Fatalf("%s: unable to bind ip from link, killing myself", cfg.NodeName)
				}
				err = fip.Bind()
				if err != nil {
					klog.Error(err)
					klog.Fatalf("%s: unable to bind ip from provider, killing myself", cfg.NodeName)
				}
				<-ctx.Done()
				klog.Infof("%s: no more leader", cfg.NodeName)
			},
			OnStoppedLeading: func() {
				err := fip.Bind()
				if err != nil {
					klog.Error(err)
					klog.Fatalf("%s: unable to unbind ip from provider, killing myself", cfg.NodeName)
				}
				err = linkIp.Bind()
				if err != nil {
					klog.Error(err)
					klog.Fatalf("%s: unable to unbind ip from link, killing myself", cfg.NodeName)
				}
				klog.Infof("%s: removing ips", cfg.NodeName)
			},
			OnNewLeader: func(i string) {
				if i != cfg.NodeName {
					klog.Infof("%s: new leader for %s is %s", cfg.NodeName, fip.String(), i)
				}
			},
		},
	)
	return err
}
