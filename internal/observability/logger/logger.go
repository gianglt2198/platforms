package oblogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	S *zap.SugaredLogger
	L *zap.Logger
}

func NewLogger(app, env string, maskFields map[string]string) (*Logger, error) {
	var (
		encoderConfig zapcore.EncoderConfig
		logLevel      zapcore.Level
	)
	switch env {
	case "prd":
		logLevel = zapcore.ErrorLevel
		encoderConfig = zap.NewProductionEncoderConfig()
	default:
		logLevel = zapcore.InfoLevel
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.LevelKey = "level"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.NameKey = "app"
	encoderConfig.EncodeName = zapcore.FullNameEncoder
	encoderConfig.CallerKey = "caller"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.MessageKey = "message"

	cfg := zap.Config{
		Encoding:          "logging",
		Level:             zap.NewAtomicLevelAt(logLevel),
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		EncoderConfig:     encoderConfig,
		DisableStacktrace: true,
	}

	if err := zap.RegisterEncoder("logging", func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return NewSensitiveFieldsEncoder(cfg, maskFields), nil
	}); err != nil {
		return nil, err
	}

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	l.Core().Sync()

	l = l.Named(app)
	zap.ReplaceGlobals(l)
	return &Logger{l.Sugar(), l}, nil
}
