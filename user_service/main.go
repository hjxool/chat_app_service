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

	r := gin.Default()

	// 公共路由（无需登录）
	public := r.Group("/api") // group传入的是前缀 可以重复
	{                         // 局部作用域块 这里是为了提升代码可读性
		public.POST("/register", auth.RegisterHandler)
		public.POST("/login", auth.LoginHandler)
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
