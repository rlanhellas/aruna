package config

import (
	"github.com/rlanhellas/aruna/global"
	"github.com/spf13/viper"
)

// HttpServerPort return http server port
func HttpServerPort() uint8 {
	return uint8(viper.GetUint(global.HttpServerPort))
}

// HttpServerEnabled return whether http server is enabled
func HttpServerEnabled() bool {
	return viper.GetBool(global.HttpServerEnabled)
}

// Custom return custom configuration
func Custom(key string) any {
	return viper.Get(key)
}
