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

// api 返回结构体
type api struct {
	JSON *response
	c    *gin.Context
}

// Success 函数用于创建一个包含成功响应的 api 对象。
// 参数 c 是 Gin 框架的上下文对象，用于处理请求和响应。
// 返回值是一个指向 API 结构体的指针，其中包含 JSON 响应数据。
func Success(c *gin.Context) *api {
	// 创建响应结构体，包含状态码和时间戳
	json := &response{
		Code:      200,               // 成功的状态码
		Timestamp: time.Now().Unix(), // 当前时间的时间戳
	}
	// 创建 api 结构体实例，并将响应数据和上下文对象赋值给它
	r := &api{
		JSON: json, // JSON 响应数据
		c:    c,    // Gin 上下文对象
	}
	return r // 返回 api 对象的指针
}

// Error 函数用于创建一个包含错误响应的 api 对象。
// 参数 c 是 Gin 框架的上下文对象，用于处理请求和响应。
// 返回值是一个指向 api 结构体的指针，其中包含 JSON 响应数据。
func Error(c *gin.Context) *api {
	json := &response{
		Code:      200,
		Timestamp: time.Now().Unix(),
	}
	r := &api{
		JSON: json,
		c:    c,
	}
	return r
}

// SetMsg 方法用于设置 api 对象的 JSON 响应数据中的 Msg 字段。
// 参数 msg 是要设置的消息字符串。
// 返回值是指向 api 结构体的指针，用于链式调用。
func (r *api) SetMsg(msg string) *api {
	r.JSON.Msg = msg
	return r
}

// SetData 方法用于设置 api 对象的 JSON 响应数据中的 Data 字段。
// 参数 data 是要设置的数据，类型为 any（任意类型）。
// 返回值是指向 api 结构体的指针，用于链式调用。
func (r *api) SetData(data any) *api {
	r.JSON.Data = data
	return r
}

// SetCode 方法用于设置 api 对象的 JSON 响应数据中的 Code 字段。
// 参数 code 是要设置的状态码，类型为 int。
// 返回值是指向 api 结构体的指针，用于链式调用。
func (r *api) SetCode(code int) *api {
	r.JSON.Code = code
	return r
}

// End 方法用于将 api 对象的 JSON 响应数据发送给客户端。
// 它使用 Gin 上下文对象的 JSON 方法将 JSON 响应数据发送给客户端。
// 参数 r 是指向 api 结构体的指针，包含 JSON 响应数据和 Gin 上下文对象。
func (r *api) End() {
	r.c.JSON(r.JSON.Code, r.JSON)
}
