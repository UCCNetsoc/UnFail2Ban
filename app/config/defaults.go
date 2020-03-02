package config

import (
	"github.com/spf13/viper"
)

func InitDefaults() {
	viper.SetDefault("http.port", "9786")

	// Consul settings
	viper.SetDefault("consul.url", "netsoc-consul:8500")
	viper.SetDefault("consul.token", "") // ACL token
	viper.SetDefault("consul.path", "windlass")

	viper.SetDefault("fail2rest.secret", "")
}
