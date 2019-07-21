package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Name is used as leaderelection namespace
	Namespace string `required:"true"`

        // NodeName is the node the pod is running on
        NodeName string `required:"true"`

	// Provider is the provide for floating ip
	Provider string `required:"true"`

	// ProviderToken is used to login to the provider
	ProviderToken string `required:"true"`

	// ProviderIPs is a list of floating ip ids
	ProviderIPs []int `required:"true"`

	// Link to bind the IPs
	Link string `required:"true"`
}

// Create parses environment variables and returns a configuration
func Create() (c *Config, err error) {
	c = &Config{}
	err = envconfig.Process("floatip", c)
	return
}
