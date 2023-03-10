package config

import (
	"github.com/rlanhellas/aruna/global"
	"github.com/spf13/viper"
)

// HttpServerPort return http server port
func HttpServerPort() int {
	return viper.GetInt(global.HttpServerPort)
}

// HttpServerEnabled return whether http server is enabled
func HttpServerEnabled() bool {
	return viper.InConfig(global.HttpServerEnabled) && viper.GetBool(global.HttpServerEnabled)
}

// LoggerLevel return logger level
func LoggerLevel() string {
	return viper.GetString(global.LoggerLevel)
}

// LoggerPath return path to write logs
func LoggerPath() string {
	return viper.GetString(global.LoggerPath)
}

// LoggerEncoding return logger encoding (console or json)
func LoggerEncoding() string {
	return viper.GetString(global.LoggerEncoding)
}

// AppName return application name
func AppName() string {
	return viper.GetString(global.AppName)
}

// AppVer return application version
func AppVer() string {
	return viper.GetString(global.AppVer)
}

// DbEnabled return whether db integration is enabled
func DbEnabled() bool {
	return viper.InConfig(global.DbEnabled) && viper.GetBool(global.DbEnabled)
}

// DbType return the db type (postgres, mysql, oracle, sqlserver, etc ...)
func DbType() string {
	return viper.GetString(global.DbType)
}

// DbConnectionString return the connection string according database type
func DbConnectionString() string {
	return viper.GetString(global.DbConnectionString)
}

// DbShowSQL return whether sql should be printed out in console
func DbShowSQL() bool {
	return viper.GetBool(global.DbShowSQL)
}

// DbSchema return the schema
func DbSchema() string {
	return viper.GetString(global.DbSchema)
}

// SecurityEnabled return whether security is enabled or not for HTTP calls
func SecurityEnabled() bool {
	return viper.InConfig(global.SecurityEnabled) && viper.GetBool(global.SecurityEnabled)
}

// SecurityClientId return clientid for OAuth2 flow
func SecurityClientId() string {
	return viper.GetString(global.SecurityClientId)
}

// SecurityClientSecret return clientsecret for OAuth2 flow
func SecurityClientSecret() string {
	return viper.GetString(global.SecurityClientSecret)
}

// SecurityTokenUri return token uri for OAuth2 flow
func SecurityTokenUri() string {
	return viper.GetString(global.SecurityTokenUri)
}

// SecurityJwkUri return jwk uri for OAuth2 flow
func SecurityJwkUri() string {
	return viper.GetString(global.SecurityJwkUri)
}

// Custom return custom configuration
func Custom(key string) any {
	return viper.Get(key)
}
