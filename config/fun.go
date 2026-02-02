package config

import (
	"fmt"
	"os"
	"path/filepath"

	"git.magicany.cc/black1552/gin-base/log"
	"git.magicany.cc/black1552/gin-base/utils"
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
		setDefault()
		err = viper.WriteConfig()
		if err != nil {
			log.Error("保存配置文件失败: ", err)
			return
		}
	}
}

func setDefault() {
	viper.Set("SERVER.addr", "127.0.0.1:8080")
	viper.Set("DATABASE.host", "127.0.0.1")
	viper.Set("DATABASE.port", 3306)
	viper.Set("DATABASE.username", "root")
	viper.Set("DATABASE.password", "")
	viper.Set("DATABASE.name", "")
	viper.Set("JWT.secret", "SET-YOUR-SECRET")
	viper.Set("JWT.expire", 86400)
}

func LoadConfigFromFile() error {
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

func SetConfigValue(key string, value any) error {
	viper.SetDefault(key, value)
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

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

func GetConfigValue(key string, def ...any) *gvar.Var {
	va := gvar.New(viper.Get(key))
	if va.IsEmpty() && len(def) > 0 {
		return gvar.New(def[0])
	}
	return va
}
func Unmarshal(s any) (any, error) {
	err := viper.Unmarshal(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func GetAllConfig() map[string]any {
	return viper.AllSettings()
}
