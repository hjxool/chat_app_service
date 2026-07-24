package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"user_service/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	jwtSecret = []byte("go_secret_key")
	db        *gorm.DB
)

// User 数据库实体模型
type User struct {
	// 数据库里 BIGINT 对应go中的int64 而叠加 UNSIGNED 无符号则变成 uint64 无符号整数类型
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;type:varchar(64);not null" json:"username"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"` // 密码 json 输出时忽略
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"` // binding 进行参数校验
	Password string `json:"password" binding:"required"`
}

type MyCustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	// 没有显式字段名的结构体 内部的字段和方法会被平铺到外层
	// 如MyCustomClaims实例就可以直接a.c取值 不用a.b.c
	// 并且自动拥有一个名为 RegisteredClaims 的字段
	// ⚠️这里虽然内部字段被提升了 但是在底层的物理结构上依然属于jwt.RegisteredClaims 这个子结构体
	// 因此初始化时 必须把它当成一个名为 RegisteredClaims 的隐式字段来赋值
	// ⚠️同时因为匿名嵌入结构体内的方法也提升到外层 因此也就继承了interface所需的方法
	// 因此满足 jwt.Claims 接口
	jwt.RegisteredClaims
}

func main() {
	// 初始化数据库连接
	initDB()

	r := gin.Default()
	// 公共路由（无需登录）
	public := r.Group("/api") // group传入的是前缀 可以重复
	{                         // 局部作用域块 这里是为了提升代码可读性
		public.POST("/login", loginHandler)
	}
	// 受保护路由（挂载 JWT 中间件）
	protected := r.Group("/api")
	protected.Use(AuthMiddleware())
	{
		protected.GET("profile", profileHandler)
	}
	fmt.Println("服务运行在8080端口...")
	r.Run(":8080") // 监听本地所有IP 冒号:不能省略
}

// 初始化 GORM MySQL 连接
func initDB() {
	// DSN 格式: root:密码@tcp(宿主机IP:端口)/数据库名?charset&parseTime&loc
	// parseTime=True MySQL 驱动自动把 DATETIME 转成 time.Time
	// loc 决定根据什么时区解析时间字符串
	dsn := `root:rootpassword@tcp(127.0.0.1:3306)/user_db?charset=utf8mb4&parseTime=True&loc=Local`
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	logger.Info("数据库连接成功！")
}

// 统一返回的响应体
type Response[T any] struct {
	Head struct { // 匿名结构体
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"head"`
	Body T `json:"body"`
}

// 统一返回成功的方法
func SuccessResponse[T any](c *gin.Context, body T) {
	res := Response[T]{}
	res.Head.Code = http.StatusOK
	res.Head.Message = "success"
	res.Body = body
	c.JSON(http.StatusOK, res)
}

// 统一返回失败的方法（不需要泛型，因为不需要 body）
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

// 登录处理器
func loginHandler(c *gin.Context) {
	var req LoginRequest
	// Gin 的 ShouldBindJSON 会自动解析 JSON，如果不合法直接报错
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "参数解析失败"+err.Error())
		return
	}
	// 根据用户名从数据库查询用户
	var user User
	result := db.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		} else {
			ErrorResponse(c, http.StatusInternalServerError, "数据库查询异常")
		}
		return
	}
	// 校验密码（利用 bcrypt 比较密文与请求传入的明文）
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 生成 Token
	// strconv.FormatUint 将Uint也就是数据库中bigint转换为10进制数字字符串
	token, err := generateToken(strconv.FormatUint(uint64(user.ID), 10), req.Username)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Token 生成失败")
		return
	}
	SuccessResponse(c, gin.H{
		"token": token,
	})
}

// 生成 Token
func generateToken(userID string, username string) (string, error) {
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

// 中间件
func AuthMiddleware() gin.HandlerFunc {
	// 闭包形式方便后续改为传参
	return func(c *gin.Context) {
		// 从请求头获取
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			ErrorResponse(c, http.StatusUnauthorized, "请求未授权，请提供 Authorization Header")
			c.Abort() // 拦截后续的处理器执行
			return    // 但当前代码块的仍会执行
		}
		// 第三个参数 表示想要返回的数组长度
		// 传小于0的数 表示完全切分
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			ErrorResponse(c, http.StatusUnauthorized, "Authorization 格式必须为 Bearer {token}")
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
			ErrorResponse(c, http.StatusUnauthorized, "Token 已失效或过期")
			c.Abort()
			return
		}
		// 把解析出来的 claims（用户信息）存入 Gin 上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next() // 放行 执行后续的 Handler
	}
}

// 受保护路由
func profileHandler(c *gin.Context) {
	// 从上下文中取出中间件注入的用户信息
	userID, _ := c.Get("userID") // 第二个返回值是exist
	username, _ := c.Get("username")
	SuccessResponse(c, gin.H{
		"user_id":  userID,
		"username": username,
		"remark":   "这是只有登录后才能看到的绝密数据",
	})
}
