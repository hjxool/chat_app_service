package common

import (
	"user_service/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化 GORM MySQL 连接
func InitDB() {
	// DSN 格式: root:密码@tcp(宿主机IP:端口)/数据库名?charset&parseTime&loc
	// parseTime=True MySQL 驱动自动把 DATETIME 转成 time.Time
	// loc 决定根据什么时区解析时间字符串
	dsn := `root:rootpassword@tcp(127.0.0.1:3306)/user_db?charset=utf8mb4&parseTime=True&loc=Local`
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	logger.Info("数据库连接成功！")
}
