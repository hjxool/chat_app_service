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
	// Gin 的 ShouldBindJSON 会按照 binding 声明校验 JSON，如果不合法直接报错
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "参数解析失败"+err.Error())
		return
	}
	// 根据用户名从数据库查询用户
	var user User
	// First 等同于 ORDER BY primary_key ASC LIMIT 1 主键升序并限制返回单条
	// 但因为定义时 Username 是不可重复值 所以 MySQL 优化器会直接忽略排序 所以不会导致慢查询
	// 但为了避免排序情况 建议用Take
	result := common.DB.Where("username = ?", req.Username).Take(&user)
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

// RegisterHandler 注册处理器
func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "参数解析失败"+err.Error())
		return
	}

	// 检查用户名是否已存在
	var existingUser User
	result := common.DB.Where("username = ?", req.Username).Take(&existingUser)
	if result.Error == nil {
		common.ErrorResponse(c, http.StatusBadRequest, "用户名已存在")
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		common.ErrorResponse(c, http.StatusInternalServerError, "数据库查询异常")
		return
	}

	// 对密码进行 bcrypt 哈希加密
	// bcrypt.DefaultCost 是计算轮数 数字越大费时越高
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		common.ErrorResponse(c, http.StatusInternalServerError, "密码加密失败")
		return
	}

	// 创建新用户记录
	newUser := User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := common.DB.Create(&newUser).Error; err != nil {
		common.ErrorResponse(c, http.StatusInternalServerError, "用户注册失败")
		return
	}
	common.SuccessResponse(c, any(nil))
}
