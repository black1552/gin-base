package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

// response 是 API 响应的数据结构体
type response struct {
	Code      int    `json:"code"`      // 状态码
	Msg       string `json:"msg"`       // 响应消息
	Data      any    `json:"data"`      // 响应数据
	Timestamp int64  `json:"timestamp"` // 响应时间戳
}

// Api 返回结构体
// Api 是 Gin 上下文的响应包装器
type Api struct {
	JSON *response    // JSON 响应数据
	c    *gin.Context // Gin 上下文对象
}

// resFile 是文件响应的数据结构体
type resFile struct {
	Code        int    `json:"code"`         // 状态码
	ContentType string `json:"content_type"` // 内容类型
	Data        []byte `json:"data"`         // 文件数据
}

// ApiFile 是 Gin 上下文的文件响应包装器
type ApiFile struct {
	JSON *resFile     // 文件响应数据
	c    *gin.Context // Gin 上下文对象
}

// SuccessFile 创建一个包含成功响应的 ApiFile 对象
// 参数 c 是 Gin 框架的上下文对象
// 返回值是指向 ApiFile 结构体的指针
func SuccessFile(c *gin.Context) *ApiFile {
	json := &resFile{
		Code: 200,
	}
	r := &ApiFile{
		JSON: json,
		c:    c,
	}
	return r
}

// SetContentType 设置文件响应的内容类型
// 参数 contentType 是要设置的 MIME 类型字符串
// 返回值是指向 ApiFile 结构体的指针，用于链式调用
func (af *ApiFile) SetContentType(contentType string) *ApiFile {
	af.JSON.ContentType = contentType
	return af
}

func (af *ApiFile) SetPng(data []byte) *ApiFile {
	af.SetContentType("image/png")
	af.SetData(data)
	return af
}

func (af *ApiFile) SetJpeg(data []byte) *ApiFile {
	af.SetContentType("image/jpeg")
	af.SetData(data)
	return af
}

func (af *ApiFile) SetGif(data []byte) *ApiFile {
	af.SetContentType("image/gif")
	af.SetData(data)
	return af
}

func (af *ApiFile) SetPdf(data []byte) *ApiFile {
	af.SetContentType("application/pdf")
	af.SetData(data)
	return af
}

func (af *ApiFile) SetXls(data []byte) *ApiFile {
	af.SetContentType("application/vnd.ms-excel")
	af.SetData(data)
	return af
}

// SetData 设置文件响应的数据
// 参数 data 是要设置的字节数据
// 返回值是指向 ApiFile 结构体的指针，用于链式调用
func (af *ApiFile) SetData(data []byte) *ApiFile {
	af.JSON.Data = data
	return af
}

// End 将文件响应数据发送给客户端
// 使用 Gin 上下文对象的 Data 方法发送原始数据
func (af *ApiFile) End() {
	af.c.Data(af.JSON.Code, af.JSON.ContentType, af.JSON.Data)
}

// Success 函数用于创建一个包含成功响应的 api 对象。
// 参数 c 是 Gin 框架的上下文对象，用于处理请求和响应。
// 返回值是一个指向 API 结构体的指针，其中包含 JSON 响应数据。
func Success(c *gin.Context) *Api {
	// 创建响应结构体，包含状态码和时间戳
	json := &response{
		Code:      200,               // 成功的状态码
		Timestamp: time.Now().Unix(), // 当前时间的时间戳
	}
	// 创建 api 结构体实例，并将响应数据和上下文对象赋值给它
	r := &Api{
		JSON: json, // JSON 响应数据
		c:    c,    // Gin 上下文对象
	}
	return r // 返回 api 对象的指针
}

// Error 创建一个包含错误响应的 Api 对象
// 参数 c 是 Gin 框架的上下文对象
// 返回值是指向 Api 结构体的指针
func Error(c *gin.Context) *Api {
	json := &response{
		Code:      200,
		Timestamp: time.Now().Unix(),
	}
	r := &Api{
		JSON: json,
		c:    c,
	}
	return r
}

// SetMsg 设置响应的消息
// 参数 msg 是要设置的消息字符串
// 返回值是指向 Api 结构体的指针，用于链式调用
func (r *Api) SetMsg(msg string) *Api {
	r.JSON.Msg = msg
	return r
}

// SetData 设置响应的数据
// 参数 data 是要设置的数据，类型为 any（任意类型）
// 返回值是指向 Api 结构体的指针，用于链式调用
func (r *Api) SetData(data any) *Api {
	r.JSON.Data = data
	return r
}

// SetCode 设置响应的状态码
// 参数 code 是要设置的状态码
// 返回值是指向 Api 结构体的指针，用于链式调用
func (r *Api) SetCode(code int) *Api {
	r.JSON.Code = code
	return r
}

// End 将 JSON 响应数据发送给客户端
// 使用 Gin 上下文对象的 JSON 方法发送响应
func (r *Api) End() {
	r.c.JSON(r.JSON.Code, r.JSON)
}
