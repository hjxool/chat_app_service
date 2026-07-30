package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"user_service/common"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("go_secret_key")

// GenerateToken 生成 Token
func GenerateToken(userID string, username string) (string, error) {
	// MyCustomClaims因为匿名嵌入结构体 因此隐式实现了Claims接口 传值类型或者指针都行
	// 但NewWithClaims只是读取 因此不用取地址
	claims := MyCustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 2)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                    // 签发时间
			Issuer:    "gin-jwt-auth",                                    // 签发人
		},
	}
	// 创建一个 Token 结构体实例（此时还没加密，只是把算法和数据组合在一起）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 传入密钥，执行签名算法，生成最终的字符串
	return token.SignedString(jwtSecret)
}

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	// 闭包形式方便后续改为传参
	return func(c *gin.Context) {
		// 从请求头获取
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.ErrorResponse(c, http.StatusUnauthorized, "请求未授权，请提供 Authorization Header")
			c.Abort() // 拦截后续的处理器执行
			return    // 但当前代码块的仍会执行
		}
		// 第三个参数 表示想要返回的数组长度
		// 传小于0的数 表示完全切分
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			common.ErrorResponse(c, http.StatusUnauthorized, "Authorization 格式必须为 Bearer {token}")
			c.Abort()
			return
		}
		tokenString := parts[1]
		// jwt.ParseWithClaims需要修改 因此需要取址传入指针
		claims := &MyCustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			// 这个回调函数的作用是 验证正在解析的token对象 通过Header["alg"]检查算法 或Header["kid"]查询密钥
			// 最后返回ParseWithClaims校验所需的密钥（因为对不同角色可能使用不同密钥加密/校验）
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				// .()类型断言 返回value, ok value是转换成目标类型后的变量 ok是断言结果
				// 这里类型断言的目的是校验收到的token是否采用与NewWithClaims(jwt.SigningMethodHS256对应的加密算法类型
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"]) // 返回error类型
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			common.ErrorResponse(c, http.StatusUnauthorized, "Token 已失效或过期")
			c.Abort()
			return
		}
		// 把解析出来的 claims（用户信息）存入 Gin 上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next() // 放行 执行后续的 Handler
	}
}
