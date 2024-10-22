package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

var (
	Logger *zap.Logger
)

// InitLogger 初始化Logger并将日志写入文件
func InitLogger() {

	dir := "log/" + bucketName + "/" + filePrefix
	os.MkdirAll(dir, 0755)
	logFile := "log/" + bucketName + "/" + filePrefix + "/" + time.Now().Format("20160102150405") + "_log.log"

	// 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	// 创建WriteSyncer，将日志写入文件
	writeSyncer := zapcore.AddSync(file)

	// 日志编码配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 构建Core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 使用JSON格式输出
		writeSyncer,
		zapcore.InfoLevel, // 日志级别
	)

	// 创建Logger
	Logger = zap.New(core)
}

// Sync 关闭日志
func Sync() {
	Logger.Sync()
}

// Info 记录info级别日志
func Info(message string, fields ...zap.Field) {
	Logger.Info(message, fields...)
}

// Error 记录error级别日志
func Error(message string, fields ...zap.Field) {
	Logger.Error(message, fields...)
}
