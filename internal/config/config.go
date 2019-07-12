package config

import (
	"github.com/kelseyhightower/envconfig"
	"k8s.io/klog"
)

type Config struct {
	// Id is used for the identifier in leaderelection
	Id string

	// Name is used as leaderelection name
	Name string `defaut:"floatip"`

	// Name is used as leaderelection namespace
	Namespace string

        // NodeName is the node the pod is running on
        NodeName string
}

// Create parses environment variables and returns a configuration
func Create() (c *Config) {
	c = &Config{}
	envconfig.Process("floatip", c)
	klog.Infof("config loaded: %+v", c)
	return
}
