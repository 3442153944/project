package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"project/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

// Init 初始化日志系统
func Init(cfg *config.Config) error {
	// 获取当前环境的日志配置
	logCfg := cfg.GetActiveLogConfig()

	// 如果日志未启用，使用 nop logger
	if !logCfg.Enabled {
		Logger = zap.NewNop()
		Sugar = Logger.Sugar()
		fmt.Printf("日志系统未启用 (mode: %s)\n", cfg.Server.Mode)
		return nil
	}

	// 确保日志目录存在
	if err := os.MkdirAll(logCfg.Path, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建多个 Core（不同级别的日志写入不同文件）
	cores := []zapcore.Core{
		createCore(logCfg, "debug.log", zapcore.DebugLevel, encoderConfig),
		createCore(logCfg, "info.log", zapcore.InfoLevel, encoderConfig),
		createCore(logCfg, "warn.log", zapcore.WarnLevel, encoderConfig),
		createCore(logCfg, "error.log", zapcore.ErrorLevel, encoderConfig),
	}

	// 如果是 debug 模式，同时输出到控制台
	if cfg.Server.Mode == "debug" {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
		cores = append(cores, consoleCore)
	}

	// 合并所有 Core
	core := zapcore.NewTee(cores...)

	// 创建 Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	Sugar = Logger.Sugar()

	fmt.Printf("日志系统初始化成功 (mode: %s, path: %s)\n", cfg.Server.Mode, logCfg.Path)
	return nil
}

// createCore 创建单个日志核心
func createCore(cfg config.LogConfig, filename string, level zapcore.Level, encoderConfig zapcore.EncoderConfig) zapcore.Core {
	// Lumberjack 配置（自动分割和清理）
	writer := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Path, filename),
		MaxSize:    cfg.FileSize,   // MB
		MaxBackups: cfg.MaxBackups, // 保留旧文件数
		MaxAge:     0,              // 不按天数删除
		Compress:   false,          // 不压缩
		LocalTime:  true,           // 使用本地时间
	}

	// 只记录当前级别的日志
	levelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l == level
	})

	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // JSON 格式
		zapcore.AddSync(writer),
		levelEnabler,
	)
}

// Sync 刷新缓冲区
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 致命错误日志（会导致程序退出）
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Panic panic 日志（会导致 panic）
func Panic(msg string, fields ...zap.Field) {
	Logger.Panic(msg, fields...)
}

// Debugf 格式化调试日志
func Debugf(template string, args ...interface{}) {
	Sugar.Debugf(template, args...)
}

// Infof 格式化信息日志
func Infof(template string, args ...interface{}) {
	Sugar.Infof(template, args...)
}

// Warnf 格式化警告日志
func Warnf(template string, args ...interface{}) {
	Sugar.Warnf(template, args...)
}

// Errorf 格式化错误日志
func Errorf(template string, args ...interface{}) {
	Sugar.Errorf(template, args...)
}

// Fatalf 格式化致命错误日志
func Fatalf(template string, args ...interface{}) {
	Sugar.Fatalf(template, args...)
}

// Panicf 格式化 panic 日志
func Panicf(template string, args ...interface{}) {
	Sugar.Panicf(template, args...)
}

// With 创建带有预设字段的子 logger
func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}
