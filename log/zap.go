package log

import (
	"os"

	"go.uber.org/zap"

	"go.uber.org/zap/zapcore"
)

type logger struct {
	sugar *zap.SugaredLogger
}

var zlog *logger

func toZapLogLevel(s string) zap.AtomicLevel {
	switch s {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "dpanic":
		return zap.NewAtomicLevelAt(zap.DPanicLevel)
	case "panic":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		panic("invalid log level " + s)
	}

}

func newLogger(level string) *logger {

	zl := toZapLogLevel(level)
	var cores []zapcore.Core

	cores = append(cores, zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		zapcore.Lock(os.Stderr),
		zl,
	))

	core := zapcore.NewTee(
		cores...,
	)
	z := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(2),
	).Sugar()

	return &logger{
		sugar: z,
	}
}

func (l *logger) Named(name string) *logger {
	return &logger{sugar: l.sugar.Named(name)}
}

func (l *logger) AddCallerSkip(skip int) *logger {
	sugar := l.sugar.Desugar().WithOptions(zap.AddCallerSkip(skip)).Sugar()
	return &logger{sugar: sugar}
}

func (l *logger) Debug(args ...any) {
	l.sugar.Debug(args...)
}
func (l *logger) Info(args ...any) {
	l.sugar.Info(args...)
}
func (l *logger) Warn(args ...any) {
	l.sugar.Warn(args...)
}
func (l *logger) Error(args ...any) {
	l.sugar.Error(args...)
}
func (l *logger) DPanic(args ...any) {
	l.sugar.DPanic(args...)
}
func (l *logger) Panic(args ...any) {
	l.sugar.Panic(args...)
}
func (l *logger) Debugf(template string, args ...any) {
	l.sugar.Debugf(template, args...)
}
func (l *logger) Infof(template string, args ...any) {
	l.sugar.Infof(template, args...)
}
func (l *logger) Warnf(template string, args ...any) {
	l.sugar.Warnf(template, args...)
}
func (l *logger) Errorf(template string, args ...any) {
	l.sugar.Errorf(template, args...)
}
func (l *logger) DPanicf(template string, args ...any) {
	l.sugar.DPanicf(template, args...)
}
func (l *logger) Panicf(template string, args ...any) {
	l.sugar.Panicf(template, args...)
}

func init() {
	zlog = newLogger("debug")
}
