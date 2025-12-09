package logging

import (
	"context"

	"go.uber.org/zap"
)

type Config struct {
	Level             string   `mapstructure:"level" json:"level" yaml:"level"`
	DisableCaller     bool     `mapstructure:"disable_caller" json:"disable_caller" yaml:"disable_caller"`
	DisableStacktrace bool     `mapstructure:"disable_stacktrace" json:"disable_stacktrace" yaml:"disable_stacktrace"`
	OutputPaths       []string `mapstructure:"output_paths" json:"output_paths" yaml:"output_paths"`
	Encoding          string   `mapstructure:"encoding" json:"encoding" yaml:"encoding"`
	UltraHuman        bool     `mapstructure:"ultra_human" json:"ultra_human" yaml:"ultra_human"`
	LogHttpRequests   bool     `mapstructure:"log_http_requests" json:"log_http_requests" yaml:"log_http_requests"`
}

type Logger interface {
	InfoCtx(ctx context.Context, msg string, fields ...zap.Field)
	WarnCtx(ctx context.Context, msg string, fields ...zap.Field)
	ErrorCtx(ctx context.Context, msg string, fields ...zap.Field)
	DebugCtx(ctx context.Context, msg string, fields ...zap.Field)
	FatalCtx(ctx context.Context, msg string, fields ...zap.Field)
	Sync() error
	Zap() *zap.Logger
	SetLevel(level string) error
	GetLevel() string
}

type zapLogger struct {
	log   *zap.Logger
	level zap.AtomicLevel
}

var global Logger

func Init(tracerName string, cfg *Config) error {
	level := zap.NewAtomicLevel()
	if cfg != nil && cfg.Level != "" {
		if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
			level.SetLevel(zap.InfoLevel)
		}
	} else {
		level.SetLevel(zap.InfoLevel)
	}
	encoding := "json"
	if cfg != nil && cfg.Encoding != "" {
		encoding = cfg.Encoding
	}
	output := []string{"stdout"}
	if cfg != nil && len(cfg.OutputPaths) > 0 {
		output = cfg.OutputPaths
	}
	zapCfg := zap.Config{
		Level:             level,
		Development:       false,
		Encoding:          encoding,
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		OutputPaths:       output,
		ErrorOutputPaths:  output,
		DisableCaller:     cfg != nil && cfg.DisableCaller,
		DisableStacktrace: cfg != nil && cfg.DisableStacktrace,
	}
	l, err := zapCfg.Build()
	if err != nil {
		return err
	}
	global = &zapLogger{log: l, level: level}
	return nil
}

func Shutdown() error {
	if global == nil {
		return nil
	}
	return global.Sync()
}

func GlobalLogger() Logger {
	return global
}

func (l *zapLogger) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.log.Info(msg, fields...)
}
func (l *zapLogger) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.log.Warn(msg, fields...)
}
func (l *zapLogger) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.log.Error(msg, fields...)
}
func (l *zapLogger) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.log.Debug(msg, fields...)
}
func (l *zapLogger) FatalCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.log.Fatal(msg, fields...)
}
func (l *zapLogger) Sync() error                 { return l.log.Sync() }
func (l *zapLogger) Zap() *zap.Logger            { return l.log }
func (l *zapLogger) SetLevel(level string) error { return l.level.UnmarshalText([]byte(level)) }
func (l *zapLogger) GetLevel() string            { return l.level.String() }

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	if global != nil {
		global.InfoCtx(ctx, msg, fields...)
	}
}
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	if global != nil {
		global.WarnCtx(ctx, msg, fields...)
	}
}
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	if global != nil {
		global.ErrorCtx(ctx, msg, fields...)
	}
}
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	if global != nil {
		global.DebugCtx(ctx, msg, fields...)
	}
}
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if global != nil {
		global.FatalCtx(ctx, msg, fields...)
	}
}
