package auth

import (
	"errors"
	"net/http"
	"user_service/common"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证路由处理器结构体
type AuthHandler struct {
	service AuthService
}

// NewAuthHandler 创建 AuthHandler 实例
func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// LoginHandler 登录处理器
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest
	// Gin 的 ShouldBindJSON 会按照 binding 声明校验 JSON，如果不合法直接报错
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "参数解析失败"+err.Error())
		return
	}

	token, err := h.service.Login(&req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrPasswordIncorrect) {
			common.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		} else {
			common.ErrorResponse(c, http.StatusInternalServerError, "数据库查询异常")
		}
		return
	}

	common.SuccessResponse(c, gin.H{
		"token": token,
	})
}

// RegisterHandler 注册处理器
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "参数解析失败"+err.Error())
		return
	}

	user, err := h.service.Register(&req)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			common.ErrorResponse(c, http.StatusInternalServerError, "用户注册失败")
		}
		return
	}

	common.SuccessResponse(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}
