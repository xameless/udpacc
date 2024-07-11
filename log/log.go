package log

func Info(args ...any) {
	zlog.Info(args...)
}

func Debug(args ...any) {
	zlog.Debug(args...)
}

func Warn(args ...any) {
	zlog.Warn(args...)
}

func Error(args ...any) {
	zlog.Error(args...)
}

func Panic(args ...any) {
	zlog.Panic(args...)
}

func Infof(template string, args ...any) {
	zlog.Infof(template, args...)
}

func Debugf(template string, args ...any) {
	zlog.Debugf(template, args...)
}

func Warnf(template string, args ...any) {
	zlog.Warnf(template, args...)
}

func Errorf(template string, args ...any) {
	zlog.Errorf(template, args...)
}

func Panicf(template string, args ...any) {
	zlog.Panicf(template, args...)
}
