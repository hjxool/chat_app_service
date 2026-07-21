package logger

import "go.uber.org/zap"

var log *zap.Logger

func Init() {
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

// 避免logger.log.fn这样的调用形式 将其封装为函数
func Sync() {
	log.Sync()
}
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}
func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}
