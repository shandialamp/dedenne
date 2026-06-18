package log

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/shandialamp/dedenne/config"
)

var logger *zap.Logger

// InitLogger 初始化全局日志
func InitLogger(cfg *config.LogConfig) error {
	var encoderCfg zapcore.EncoderConfig

	if cfg.Format == "json" {
		encoderCfg = zap.NewProductionEncoderConfig()
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	// 确定日志输出位置
	var syncer zapcore.WriteSyncer
	switch cfg.OutputPath {
	case "stdout":
		syncer = zapcore.AddSync(os.Stdout)
	case "stderr":
		syncer = zapcore.AddSync(os.Stderr)
	default:
		// 文件路径
		file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		syncer = zapcore.AddSync(file)
	}

	// 转换日志级别
	level := parseLogLevel(cfg.Level)

	core := zapcore.NewCore(encoder, syncer, level)
	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// L 获取全局日志记录器
func L() *zap.Logger {
	if logger == nil {
		panic("logger not initialized, call InitLogger first")
	}
	return logger
}

// Sync 同步日志缓冲到磁盘
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
