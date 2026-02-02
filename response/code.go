package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

type response struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Data      any    `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

type Api struct {
	Json *response
	c    *gin.Context
}

func Success(c *gin.Context) *Api {
	json := &response{
		Code:      200,
		Timestamp: time.Now().Unix(),
	}
	r := &Api{
		Json: json,
		c:    c,
	}
	return r
}
func Error(c *gin.Context) *Api {
	json := &response{
		Code:      200,
		Timestamp: time.Now().Unix(),
	}
	r := &Api{
		Json: json,
		c:    c,
	}
	return r
}
func (r *Api) SetMsg(msg string) *Api {
	r.Json.Msg = msg
	return r
}
func (r *Api) SetData(data any) *Api {
	r.Json.Data = data
	return r
}
func (r *Api) SetCode(code int) *Api {
	r.Json.Code = code
	return r
}
func (r *Api) End() {
	r.c.JSON(r.Json.Code, r.Json)
}
