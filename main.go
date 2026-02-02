package main

import (
	"git.magicany.cc/black1552/gin-base/config"
	"git.magicany.cc/black1552/gin-base/log"
	"git.magicany.cc/black1552/gin-base/response"
	"github.com/gin-gonic/gin"
)

// TIP <p>To run your response, right-click the response and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	err := config.LoadConfigFromFile()
	if err != nil {
		log.Info("err: ", err.Error())
	}
	err = config.SetConfigValue("SERVICE.user-service", "127.0.0.1:3001")
	if err != nil {
		log.Info("err: ", err.Error())
	}
	err = config.SetConfigMap(map[string]any{
		"SERVICE.product-service": "127.0.0.1:3002",
		"SERVICE.order-service":   "127.0.0.1:3003",
	})
	response.Success(&gin.Context{}).SetMsg("text").End()
	log.Info("ceshi", config.GetAllConfig())
}
