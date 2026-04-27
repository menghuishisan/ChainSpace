package logger

import (
	"os"
	"path/filepath"

	"github.com/chainspace/backend/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger
var sugar *zap.SugaredLogger

// Init 初始化日志
func Init(cfg *config.LogConfig) error {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "file" {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writeSyncer = zapcore.AddSync(file)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = log.Sugar()

	return nil
}

// L 返回zap.Logger
func L() *zap.Logger {
	if log == nil {
		log, _ = zap.NewDevelopment()
	}
	return log
}

// S 返回SugaredLogger
func S() *zap.SugaredLogger {
	if sugar == nil {
		l, _ := zap.NewDevelopment()
		sugar = l.Sugar()
	}
	return sugar
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

// Debugf 调试日志（格式化）
func Debugf(template string, args ...interface{}) {
	S().Debugf(template, args...)
}

// Infof 信息日志（格式化）
func Infof(template string, args ...interface{}) {
	S().Infof(template, args...)
}

// Warnf 警告日志（格式化）
func Warnf(template string, args ...interface{}) {
	S().Warnf(template, args...)
}

// Errorf 错误日志（格式化）
func Errorf(template string, args ...interface{}) {
	S().Errorf(template, args...)
}

// Fatalf 致命错误日志（格式化）
func Fatalf(template string, args ...interface{}) {
	S().Fatalf(template, args...)
}

// With 添加字段
func With(fields ...zap.Field) *zap.Logger {
	return L().With(fields...)
}

// Sync 刷新日志缓冲
func Sync() error {
	if log != nil {
		return log.Sync()
	}
	return nil
}
