package common

import (
	"context"
	"user_service/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var RDB *redis.Client

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // docker-compose 没设密码
		// Redis 默认16个逻辑数据库 但不建议指定DB 因为集群模式下仅支持使用 DB 0不支持多数据库切换
		// 更推荐通过 Key 前缀 如 app:user:1 做业务隔离，而非依赖不同 DB 编号
		// DB: 0,
	})

	// 检测连接
	// 传入context是为了方便链路追踪 此时还没有请求所以只能用永不取消、永不超时的context.Background()
	// 其实go 1.21+版本后推荐用 asyncCtx := context.WithoutCancel(ctx) 代替
	if err := RDB.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("Redis 连接失败", zap.Error(err))
	}
	logger.Info("Redis 连接成功！")
}
