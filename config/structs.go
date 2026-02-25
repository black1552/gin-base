package config

// BaseConfig 基础配置信息
type BaseConfig struct {
	Server   ServerConfig   `mapstructure:"SERVER"`
	Database DataBaseConfig `mapstructure:"DATABASE"`
	Jwt      JwtConfig      `mapstructure:"JWT"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Addr string `mapstructure:"addr"`
	Mode string `mapstructure:"mode"`
}

// DataBaseConfig 数据库配置
type DataBaseConfig struct {
	Dns  string `mapstructure:"dns"`
	Type string `mapstructure:"type"`
}

// JwtConfig JWT配置
type JwtConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int64  `mapstructure:"expire"`
}
