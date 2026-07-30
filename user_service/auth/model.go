package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// LoginRequest 登录请求参数结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // binding 进行参数校验
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求参数结构体
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// MyCustomClaims JWT 自定义声明
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
