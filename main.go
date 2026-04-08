package main

import (
	"git.magicany.cc/black1552/gin-base/config"
	"git.magicany.cc/black1552/gin-base/log"
	"git.magicany.cc/black1552/gin-base/server"
)

// TIP <p>To run your response, right-click the response and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	g := server.New()
	cf, err := config.Unmarshal[config.BaseConfig]()
	if err != nil {
		log.Error("转换配置失败", err)
	}
	log.Info("启动服务:", cf.Server.Addr)
	server.Run(g)
}
