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

// SetDefault 设置默认配置信息
func SetDefault() {
	viper.Set("SERVER.addr", "127.0.0.1:8080")
	viper.Set("SERVER.mode", "release")
	viper.Set("DATABASE.type", "sqlite")
	viper.Set("DATABASE.dns", gfile.Join(gfile.Pwd(), "db", "database.db"))
	viper.Set("DATABASE.debug", true)
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

// GetAllConfig 获取所有配置信息并返回Map
func GetAllConfig() map[string]any {
	return viper.AllSettings()
}
