package config

type BaseConfig struct {
	Server   ServerConfig   `mapstructure:"SERVER"`
	Database DataBaseConfig `mapstructure:"DATABASE"`
	Jwt      JwtConfig      `mapstructure:"JWT"`
}
type ServerConfig struct {
	Addr string `mapstructure:"addr"`
	Mode string `mapstructure:"mode"`
}

type DataBaseConfig struct {
	Dns  string `mapstructure:"dns"`
	Type string `mapstructure:"type"`
}

type JwtConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int64  `mapstructure:"expire"`
}
