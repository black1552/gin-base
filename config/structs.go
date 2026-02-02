package config

type BaseConfig struct {
	Server   ServerConfig   `toml:"SERVER"`
	Database DataBaseConfig `toml:"DATABASE"`
	Jwt      JwtConfig      `toml:"JWT"`
	Logger   Logger         `toml:"LOGGER"`
}
type ServerConfig struct {
	Addr string `toml:"addr"`
}

type DataBaseConfig struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
	User string `toml:"user"`
	Pwd  string `toml:"pwd"`
	Name string `toml:"name"`
}

type JwtConfig struct {
	Secret string `toml:"secret"`
	Expire int64  `toml:"expire"`
}

type Logger struct {
	Level string `toml:"level"`
	Path  string `toml:"path"`
}
