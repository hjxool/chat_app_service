package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一返回的响应体
type Response[T any] struct {
	Head struct { // 匿名结构体
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"head"`
	Body T `json:"body"`
}

// SuccessResponse 统一返回成功的方法
func SuccessResponse[T any](c *gin.Context, body T) {
	res := Response[T]{}
	res.Head.Code = http.StatusOK
	res.Head.Message = "success"
	res.Body = body
	c.JSON(http.StatusOK, res)
}

// ErrorResponse 统一返回失败的方法（不需要泛型，因为不需要 body）
func ErrorResponse(c *gin.Context, code int, msg string) {
	res := gin.H{
		"head": gin.H{
			"code":    code,
			"message": msg,
		},
		"body": nil,
	}
	c.JSON(http.StatusOK, res)
}
