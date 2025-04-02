// 日志模块，用于IoC容器的日志记录
package ioc

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 日志级别常量
const (
	// DebugLevel 调试级别，详细的开发信息
	DebugLevel = "debug"
	// InfoLevel 信息级别，常规操作信息
	InfoLevel = "info"
	// WarnLevel 警告级别，可能的问题
	WarnLevel = "warn"
	// ErrorLevel 错误级别，影响功能但不致命的错误
	ErrorLevel = "error"
	// FatalLevel 致命级别，导致程序终止的错误
	FatalLevel = "fatal"
)

// Logger 日志接口
type Logger interface {
	// Debug 记录调试级别日志
	Debug(msg string, fields ...zapcore.Field)
	// Info 记录信息级别日志
	Info(msg string, fields ...zapcore.Field)
	// Warn 记录警告级别日志
	Warn(msg string, fields ...zapcore.Field)
	// Error 记录错误级别日志
	Error(msg string, fields ...zapcore.Field)
	// Fatal 记录致命级别日志并退出程序
	Fatal(msg string, fields ...zapcore.Field)
	// With 返回带有指定字段的新日志记录器
	With(fields ...zapcore.Field) Logger
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	// 日志级别，默认info
	Level string `json:"level"`
	// 是否使用JSON格式输出，默认false
	EnableJSON bool `json:"enableJson"`
	// 是否输出到文件，默认false
	OutputFile bool `json:"outputFile"`
	// 日志文件路径，默认./logs/ioc.log
	FilePath string `json:"filePath"`
	// 是否输出到控制台，默认true
	OutputConsole bool `json:"outputConsole"`
	// 是否开启调用者信息，默认true
	EnableCaller bool `json:"enableCaller"`
	// 是否开发模式，默认false
	Development bool `json:"development"`
}

// loggerImpl 日志实现
type loggerImpl struct {
	logger *zap.Logger
}

var (
	// 默认日志实例
	defaultLogger Logger
	// 日志初始化锁
	loggerOnce sync.Once
	// 默认配置
	defaultConfig = LoggerConfig{
		Level:         InfoLevel,
		EnableJSON:    false,
		OutputFile:    false,
		FilePath:      "./logs/ioc.log",
		OutputConsole: true,
		EnableCaller:  true,
		Development:   false,
	}
)

// 初始化日志
func initLogger(config LoggerConfig) {
	var level zapcore.Level
	switch config.Level {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	case FatalLevel:
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// 编码器配置
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

	if config.Development {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 创建输出
	var cores []zapcore.Core

	// 控制台输出
	if config.OutputConsole {
		var encoder zapcore.Encoder
		if config.EnableJSON {
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		} else {
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		}
		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			zap.NewAtomicLevelAt(level),
		)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if config.OutputFile && config.FilePath != "" {
		// 确保日志目录存在
		dir := config.FilePath[:len(config.FilePath)-len("/ioc.log")]
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
		}

		// 打开日志文件
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
			fileCore := zapcore.NewCore(
				fileEncoder,
				zapcore.AddSync(file),
				zap.NewAtomicLevelAt(level),
			)
			cores = append(cores, fileCore)
		}
	}

	// 合并cores
	core := zapcore.NewTee(cores...)

	// 创建Logger
	var logger *zap.Logger
	if config.EnableCaller {
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		logger = zap.New(core)
	}

	// 设置开发模式
	if config.Development {
		logger = logger.WithOptions(zap.Development())
	}

	defaultLogger = &loggerImpl{
		logger: logger,
	}

	defaultLogger.Info("IoC容器日志系统初始化完成",
		zap.String("level", config.Level),
		zap.Bool("json", config.EnableJSON),
		zap.Bool("file", config.OutputFile),
		zap.Bool("console", config.OutputConsole),
		zap.Bool("dev", config.Development))
}

// GetLogger 获取日志实例
func GetLogger() Logger {
	loggerOnce.Do(func() {
		initLogger(defaultConfig)
	})
	return defaultLogger
}

// ConfigureLogger 配置日志系统
func ConfigureLogger(config LoggerConfig) {
	loggerOnce.Do(func() {
		initLogger(config)
	})
}

// Debug 实现Debug级别日志
func (l *loggerImpl) Debug(msg string, fields ...zapcore.Field) {
	l.logger.Debug(msg, fields...)
}

// Info 实现Info级别日志
func (l *loggerImpl) Info(msg string, fields ...zapcore.Field) {
	l.logger.Info(msg, fields...)
}

// Warn 实现Warn级别日志
func (l *loggerImpl) Warn(msg string, fields ...zapcore.Field) {
	l.logger.Warn(msg, fields...)
}

// Error 实现Error级别日志
func (l *loggerImpl) Error(msg string, fields ...zapcore.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal 实现Fatal级别日志
func (l *loggerImpl) Fatal(msg string, fields ...zapcore.Field) {
	l.logger.Fatal(msg, fields...)
}

// With 返回带有指定字段的新日志记录器
func (l *loggerImpl) With(fields ...zapcore.Field) Logger {
	return &loggerImpl{
		logger: l.logger.With(fields...),
	}
}
