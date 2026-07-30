package auth

import (
	"net/http"
	"strconv"
	"user_service/common"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginHandler 登录处理器
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	// Gin 的 ShouldBindJSON 会自动解析 JSON，如果不合法直接报错
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "参数解析失败"+err.Error())
		return
	}
	// 根据用户名从数据库查询用户
	var user User
	result := common.DB.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			common.ErrorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		} else {
			common.ErrorResponse(c, http.StatusInternalServerError, "数据库查询异常")
		}
		return
	}
	// 校验密码（利用 bcrypt 比较密文与请求传入的明文）
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		common.ErrorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 生成 Token
	// strconv.FormatUint 将Uint也就是数据库中bigint转换为10进制数字字符串
	token, err := GenerateToken(strconv.FormatUint(user.ID, 10), req.Username)
	if err != nil {
		common.ErrorResponse(c, http.StatusInternalServerError, "Token 生成失败")
		return
	}
	common.SuccessResponse(c, gin.H{
		"token": token,
	})
}
