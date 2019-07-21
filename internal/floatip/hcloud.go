package floatip

import (
	"context"
	"encoding/base32"
	"fmt"

	"github.com/foosinn/kube-floatip/internal/config"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type (
	// HCloud provides floating up for Hetzner
	HCloud struct {
		ctx    context.Context
		config *config.Config
		client *hcloud.Client
		server *hcloud.Server
		ip     *hcloud.FloatingIP
	}
)

func NewHCloud(ipID int, c *config.Config) (*HCloud, error) {
	ctx := context.Background()
	client := hcloud.NewClient(hcloud.WithToken(c.ProviderToken))
	server, _, err := client.Server.GetByName(ctx, c.NodeName)
	if err != nil {
		return nil, fmt.Errorf("unable to find server: %s", err)
	}

	ip, _, err := client.FloatingIP.GetByID(ctx, ipID)
	if err != nil {
		return nil, fmt.Errorf("unable to find floatingip %d: %s", ipID, err)
	}
	return &HCloud{
		ctx:    ctx,
		config: c,
		client: client,
		server: server,
		ip:     ip,
	}, nil
}

func (h *HCloud) Bind() (err error) {
	ctx := context.Background()
	_, _, err = h.client.FloatingIP.Assign(ctx, h.ip, h.server)
	if err != nil {
		return fmt.Errorf("unable to bind ip: %s", err)
	}
	return nil
}

func (h *HCloud) Unbind() (err error) {
	return nil
}

func (h *HCloud) DnsName() string {
	e := base32.NewEncoding("abcdefghijklmnopqrstuvwxyz012345")
	e = e.WithPadding(base32.NoPadding)
	return e.EncodeToString(h.ip.IP)
}

func (h *HCloud) String() string {
	return h.ip.IP.String()
}
