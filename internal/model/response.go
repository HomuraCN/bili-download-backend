package model

// Result 对应 Java 中的 Result<T> 泛型类
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"` // interface{} 类似于 Java 的 Object，可以放任何类型
}

// ResultEnum 对应 Java 中的 ResultEnum 枚举
// Go 没有像 Java 那样复杂的枚举，通常用 const 定义
const (
	CodeSuccess = 200
	CodeError   = 500
)

// Success 成功响应
func Success(data interface{}) Result {
	return Result{
		Code: CodeSuccess,
		Msg:  "success",
		Data: data,
	}
}

// Fail 失败响应
func Fail(msg string) Result {
	return Result{
		Code: CodeError,
		Msg:  msg,
		Data: nil,
	}
}
