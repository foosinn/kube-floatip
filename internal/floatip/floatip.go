package floatip

import (
	"context"
	"github.com/foosinn/kube-floatip/internal/config"
)

type (
	// Floatip is an interface to assing floating ips
	FloatIP interface {
		Bind(context.Context) (error)
		DnsName() string
                String() string
	}
)

func New(ipID int, c *config.Config) (f FloatIP, err error){
	switch c.Provider {
	case "hcloud":
		f, err = NewHCloud(ipID, c)
	}
	return
}
