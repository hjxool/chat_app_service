package main

import (
	"fmt"
	"user_service/auth"
	"user_service/common"
	"user_service/logger"
	"user_service/profile"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志系统
	logger.Init()
	// 初始化数据库连接
	common.InitDB()
	// 初始化Redis连接
	common.InitRedis()

	// 分层初始化：Repository -> Service -> Handler
	// 这样解耦的好处是替换源或者Mock时不需要每次都NewUserRepository...
	// 可以从任意层创建结构体 实现对应的接口 将其传入即可使用
	userRepo := auth.NewUserRepository(common.DB)
	tokenRepo := auth.NewRedisTokenRepository(common.RDB)
	authService := auth.NewAuthService(userRepo, tokenRepo)
	authHandler := auth.NewAuthHandler(authService)

	r := gin.Default()

	// 公共路由（无需登录）
	public := r.Group("/api") // group传入的是前缀 可以重复
	{                         // 局部作用域块 这里是为了提升代码可读性
		public.POST("/register", authHandler.RegisterHandler)
		public.POST("/login", authHandler.LoginHandler)
	}

	// 受保护路由（挂载 JWT 中间件）
	protected := r.Group("/api")
	protected.Use(auth.AuthMiddleware())
	{
		protected.GET("profile", profile.ProfileHandler)
	}

	fmt.Println("服务运行在8080端口...")
	r.Run(":8080") // 监听本地所有IP 冒号:不能省略
}
