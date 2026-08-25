package auth

import (
	"errors"
	"net/http"
	"user_service/common"

	"github.com/gin-gonic/gin"
)

// 认证路由处理器结构体
type AuthHandler struct {
	service AuthService
}

// 创建 AuthHandler 实例
func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// 登录处理器
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest
	// Gin 的 ShouldBindJSON 会按照 binding 声明校验 JSON，如果不合法直接报错
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "参数解析失败"+err.Error())
		return
	}

	token, err := h.service.Login(c.Request.Context(), &req)
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

// 注册处理器
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

// 登出处理器
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	// 中间件已经将token解码取出字段存入gin.Context
	userID, exists := c.Get("userID")
	if !exists {
		common.ErrorResponse(c, http.StatusInternalServerError, "登出失败")
		return
	}
	// Logout 要传入string c.Get获取的是any类型 因此要类型断言
	if err := h.service.Logout(c.Request.Context(), userID.(string)); err != nil {
		common.ErrorResponse(c, http.StatusInternalServerError, "登出失败")
		return
	}
	common.SuccessResponse[any](c, nil)
}

// 验证码处理器
func (h *AuthHandler) SendCodeHandler(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "参数解析失败: "+err.Error())
		return
	}
	if err := h.service.SendVerifyCode(c.Request.Context(), &req); err != nil {
		common.ErrorResponse(c, http.StatusInternalServerError, "发送失败: "+err.Error())
		return
	}
	common.SuccessResponse[any](c, nil)
}
func (h *AuthHandler) VerifyCodeHandler(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "参数解析失败: "+err.Error())
		return
	}
	if err := h.service.VerifyCode(c.Request.Context(), &req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	common.SuccessResponse[any](c, nil)
}
