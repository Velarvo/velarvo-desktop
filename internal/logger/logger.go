package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	Level       string
	Development bool
}

var (
	mu     sync.RWMutex
	Log    = zap.NewNop().Sugar()
	base   = zap.NewNop()
	closed bool
)

func Init(opts Options) error {
	level, err := parseLevel(opts.Level)
	if err != nil {
		return err
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		MessageKey:    "msg",
		NameKey:       "logger",
		CallerKey:     "caller",
		StacktraceKey: "stacktrace",

		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   timeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	if opts.Development {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		level,
	)

	instance := zap.New(
		core,
		zap.AddCaller(),
	)

	mu.Lock()
	defer mu.Unlock()

	base = instance
	Log = instance.Sugar()
	closed = false

	return nil
}

func MustInit(opts Options) {
	if err := Init(opts); err != nil {
		panic(err)
	}
}

func Default() *zap.SugaredLogger {
	mu.RLock()
	defer mu.RUnlock()
	return Log
}

func Named(name string) *zap.SugaredLogger {
	mu.RLock()
	defer mu.RUnlock()
	return base.Named(name).Sugar()
}

func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if closed {
		return nil
	}

	if Log == nil {
		closed = true
		return nil
	}

	if err := Log.Sync(); err != nil {
		if isIgnorableSyncError(err) {
			closed = true
			return nil
		}
		return err
	}

	closed = true
	return nil
}

func parseLevel(value string) (zapcore.Level, error) {
	if strings.TrimSpace(value) == "" {
		return zapcore.InfoLevel, nil
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(value)))); err != nil {
		return 0, fmt.Errorf("invalid log level %q: %w", value, err)
	}

	return level, nil
}

func isIgnorableSyncError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "invalid argument") || strings.Contains(message, "bad file descriptor")
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05"))
}
