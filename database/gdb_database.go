package database

import (
	"context"
	"fmt"

	"github.com/black1552/gin-base/config"
	"github.com/black1552/gin-base/database/instance"
	"github.com/black1552/gin-base/database/intlog"
	customLog "github.com/black1552/gin-base/log"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gutil"
)

const (
	frameCoreComponentNameDatabase = "core.component.database"
	ConfigNodeNameDatabase         = "database"
)

func Database(name ...string) DB {
	var (
		ctx   = context.Background()
		group = gdb.DefaultGroupName
	)

	if len(name) > 0 && name[0] != "" {
		group = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", frameCoreComponentNameDatabase, group)
	db := instance.GetOrSetFuncLock(instanceKey, func() interface{} {
		// It ignores returned error to avoid file no found error while it's not necessary.
		var (
			configMap     map[string]interface{}
			configNodeKey = ConfigNodeNameDatabase
		)
		// It firstly searches the configuration of the instance name.
		if configData := config.GetAllConfig(); len(configData) > 0 {
			if v, _ := gutil.MapPossibleItemByKey(configData, ConfigNodeNameDatabase); v != "" {
				configNodeKey = v
			}
		}
		if v := config.GetConfigValue(configNodeKey); !v.IsEmpty() {
			configMap = v.Map()
		}
		// 检查配置是否存在
		if len(configMap) == 0 {
			// 从 config 包获取所有配置，检查是否包含数据库配置
			allConfig := config.GetAllConfig()
			if len(allConfig) == 0 {
				panic(gerror.NewCodef(
					gcode.CodeMissingConfiguration,
					`database initialization failed: configuration file is empty or invalid`,
				))
			}
			// 检查是否配置了数据库节点
			if _, exists := allConfig["database"]; !exists {
				// 尝试其他可能的键名
				found := false
				for key := range allConfig {
					if key == "DATABASE" || key == "Database" {
						found = true
						break
					}
				}
				if !found {
					panic(gerror.NewCodef(
						gcode.CodeMissingConfiguration,
						`database initialization failed: configuration missing for database node "%s"`,
						ConfigNodeNameDatabase,
					))
				}
			}
		}

		if len(configMap) == 0 {
			configMap = make(map[string]interface{})
		}
		// Parse `m` as map-slice and adds it to global configurations for package gdb.
		for g, groupConfig := range configMap {
			cg := ConfigGroup{}
			switch value := groupConfig.(type) {
			case []interface{}:
				for _, v := range value {
					if node := parseDBConfigNode(v); node != nil {
						cg = append(cg, *node)
					}
				}
			case map[string]interface{}:
				if node := parseDBConfigNode(value); node != nil {
					cg = append(cg, *node)
				}
			}
			if len(cg) > 0 {
				if GetConfig(group) == nil {
					intlog.Printf(ctx, "add configuration for group: %s, %#v", g, cg)
					SetConfigGroup(g, cg)
				} else {
					intlog.Printf(ctx, "ignore configuration as it already exists for group: %s, %#v", g, cg)
					intlog.Printf(ctx, "%s, %#v", g, cg)
				}
			}
		}
		// Parse `m` as a single node configuration,
		// which is the default group configuration.
		if node := parseDBConfigNode(configMap); node != nil {
			cg := ConfigGroup{}
			if node.Link != "" || node.Host != "" {
				cg = append(cg, *node)
			}
			if len(cg) > 0 {
				if GetConfig(group) == nil {
					intlog.Printf(ctx, "add configuration for group: %s, %#v", DefaultGroupName, cg)
					SetConfigGroup(DefaultGroupName, cg)
				} else {
					intlog.Printf(
						ctx,
						"ignore configuration as it already exists for group: %s, %#v",
						DefaultGroupName, cg,
					)
					intlog.Printf(ctx, "%s, %#v", DefaultGroupName, cg)
				}
			}
		}

		// Create a new ORM object with given configurations.
		if db, err := NewByGroup(name...); err == nil {
			// 自动初始化自定义日志器并设置到数据库
			db.SetLogger(customLog.GetLogger())
			return db
		} else {
			// If panics, often because it does not find its configuration for given group.
			panic(err)
		}
		return nil
	})
	if db != nil {
		return db.(DB)
	}
	return nil
}

func parseDBConfigNode(value interface{}) *ConfigNode {
	nodeMap, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}
	var (
		node = &ConfigNode{}
		err  = gconv.Struct(nodeMap, node)
	)
	if err != nil {
		panic(err)
	}
	// Find possible `Link` configuration content.
	if _, v := gutil.MapPossibleItemByKey(nodeMap, "Link"); v != nil {
		node.Link = gconv.String(v)
	}
	return node
}
