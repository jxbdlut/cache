package logging

import (
	"go.uber.org/zap"
	"time"
	"go.uber.org/zap/zapcore"
)

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + t.Format("2006-01-02 15:04:05.000") + "]")
}

func Logger(l zapcore.Level) *zap.SugaredLogger {
	encoderCfg := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	logLevel := zap.NewAtomicLevelAt(l)
	logCfg := zap.Config{
		Level:            logLevel,
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    encoderCfg,
		DisableStacktrace: true,
		OutputPaths:      []string{"stdout", "./logs/orm.log"},
		ErrorOutputPaths: []string{"stdout", "./logs/orm.log"},
	}
	logger, _ := logCfg.Build()
	return logger.Sugar()
}