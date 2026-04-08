package config

import (
	"fmt"
	"os"
	"path/filepath"

	"git.magicany.cc/black1552/gin-base/log"
	"git.magicany.cc/black1552/gin-base/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/spf13/viper"
)

// 全局配置变量
var (
	configPath string
)

func init() {
	path, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("获取当前目录失败: %v", err))
	}
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath(filepath.Join(path, "config"))
	viper.WatchConfig()
	viper.AutomaticEnv()
	configPath = filepath.Join(path, "config", "config.toml")
	if !utils.FileExists(configPath) {
		_, err = gfile.Create(configPath)
		if err != nil {
			log.Error("创建配置文件失败: ", err)
			return
		}
		log.Info("配置文件是否为空", utils.EmptyFile(configPath))
		SetDefault()
		err = viper.WriteConfig()
		if err != nil {
			log.Error("保存配置文件失败: ", err)
			return
		}
	} else {
		err = viper.ReadInConfig()
		if err != nil {
			log.Error("读取配置文件失败: ", err)
			return
		}
	}
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Info("配置文件已修改")
	})
}

func GetConfigPath() string {
	return configPath
}

// SetDefault 设置默认配置信息
func SetDefault() {
	viper.Set("SERVER.addr", "127.0.0.1:8080")
	viper.Set("SERVER.mode", "release")
	viper.Set("DATABASE.default.host", "127.0.0.1")
	viper.Set("DATABASE.default.port", "3306")
	viper.Set("DATABASE.default.user", "root")
	viper.Set("DATABASE.default.pass", "123456")
	viper.Set("DATABASE.default.name", "test")
	viper.Set("DATABASE.default.type", "mysql")
	viper.Set("DATABASE.default.role", "master")
	viper.Set("DATABASE.default.debug", false)
	viper.Set("DATABASE.default.prefix", "")
	viper.Set("DATABASE.default.dryRun", false)
	viper.Set("DATABASE.default.charset", "utf8")
	viper.Set("DATABASE.default.timezone", "Local")
	viper.Set("DATABASE.default.createdAt", "create_time")
	viper.Set("DATABASE.default.updatedAt", "update_time")
	viper.Set("DATABASE.default.timeMaintainDisabled", false)

	// JWT 配置
	viper.Set("JWT.secret", "SET-YOUR-SECRET")
	viper.Set("JWT.expire", 86400)
}

// LoadConfigFromFile 在配置文件中加载配置
func LoadConfigFromFile() error {
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

// SetConfigValue 设置指定配置文件的值
func SetConfigValue(key string, value any) error {
	viper.SetDefault(key, value)
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

// SetConfigMap 使用Map的方式添加配置信息
func SetConfigMap(value map[string]any) error {
	if len(value) == 0 {
		log.Error("value is empty")
		return fmt.Errorf("value is empty")
	}
	for k, v := range value {
		viper.SetDefault(k, v)
	}
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

// GetConfigValue  获取指定配置的值
func GetConfigValue(key string, def ...any) *gvar.Var {
	value := gvar.New(viper.Get(key))
	if value.IsEmpty() && len(def) > 0 {
		return gvar.New(def[0])
	}
	return value
}

// Unmarshal 将配置解析成指定的对象
func Unmarshal[T any]() (*T, error) {
	var s T
	err := viper.Unmarshal(&s)
	return &s, err
}

// GetAllConfig 获取所有配置信息并返回 Map
func GetAllConfig() map[string]any {
	return viper.AllSettings()
}

// GetDatabaseConfig 获取数据库配置信息
func GetDatabaseConfig() map[string]any {
	return map[string]any{
		"type":            GetConfigValue("DATABASE.type", "sqlite").String(),
		"dns":             GetConfigValue("DATABASE.dns", "").String(),
		"debug":           GetConfigValue("DATABASE.debug", true).Bool(),
		"maxIdleConns":    GetConfigValue("DATABASE.maxIdleConns", 10).Int(),
		"maxOpenConns":    GetConfigValue("DATABASE.maxOpenConns", 100).Int(),
		"connMaxLifetime": GetConfigValue("DATABASE.connMaxLifetime", 3600).Int(),
		"replicas":        GetConfigValue("DATABASE.replicas", []string{}).Strings(),
		"readPolicy":      GetConfigValue("DATABASE.readPolicy", "random").String(),
	}
}

// GetDatabaseTimeConfig 获取数据库时间配置
func GetDatabaseTimeConfig() map[string]string {
	return map[string]string{
		"createdAt": GetConfigValue("DATABASE.timeConfig.createdAt", "created_at").String(),
		"updatedAt": GetConfigValue("DATABASE.timeConfig.updatedAt", "updated_at").String(),
		"deletedAt": GetConfigValue("DATABASE.timeConfig.deletedAt", "deleted_at").String(),
		"format":    GetConfigValue("DATABASE.timeConfig.format", "2006-01-02 15:04:05").String(),
	}
}

// GetServerConfig 获取服务器配置信息
func GetServerConfig() map[string]string {
	return map[string]string{
		"addr": GetConfigValue("SERVER.addr", "127.0.0.1:8080").String(),
		"mode": GetConfigValue("SERVER.mode", "release").String(),
	}
}

// GetJWTConfig 获取 JWT 配置信息
func GetJWTConfig() map[string]any {
	return map[string]any{
		"secret": GetConfigValue("JWT.secret", "SET-YOUR-SECRET").String(),
		"expire": GetConfigValue("JWT.expire", 86400).Int(),
	}
}
