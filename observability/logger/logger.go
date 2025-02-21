package oblogger

import (
	"context"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	obLogger struct {
		S *zap.SugaredLogger
		L *zap.Logger
	}

	ObLogger interface {
		GetLogger() *zap.Logger
		GetSugaredLogger() *zap.SugaredLogger
		Info(ctx context.Context, message string, params ...interface{})
		Error(ctx context.Context, message string, err interface{})
		Warn(ctx context.Context, message string, params ...interface{})
		InfoWithMask(ctx context.Context, message, secret string)
	}
)

func NewLogger(isProdEnv bool) *obLogger {

	var coreArr []zapcore.Core

	// Log levels
	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { // Error level
		return lev >= zap.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { // Info and debug levels, debug level is the lowest
		return lev < zap.ErrorLevel && lev >= zap.DebugLevel
	})

	if isProdEnv {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder := zapcore.NewJSONEncoder(encoderConfig)

		// Info file writeSyncer
		infoFileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "./log/info.log", // Log file storage directory. If the folder does not exist, it will be created automatically.
			MaxSize:    2,                // File size limit, unit MB
			MaxBackups: 100,              // Maximum number of retained log files
			MaxAge:     30,               // Number of days to retain log files
			Compress:   true,             // Whether to compress
		})
		infoFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(infoFileWriteSyncer), lowPriority) // The third and subsequent parameters are the log levels for writing to the file. In ErrorLevel mode, only error - level logs are recorded.
		// Error file writeSyncer
		errorFileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "./log/error.log", // Log file storage directory
			MaxSize:    1,                 // File size limit, unit MB
			MaxBackups: 5,                 // Maximum number of retained log files
			MaxAge:     30,                // Number of days to retain log files
			Compress:   true,              // Whether to compress
		})
		errorFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(errorFileWriteSyncer), highPriority) // The third and subsequent parameters are the log levels for writing to the file. In ErrorLevel mode, only error - level logs are recorded.

		coreArr = append(coreArr, infoFileCore)
		coreArr = append(coreArr, errorFileCore)

		// infoLogCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), lowPriority)   // The third and subsequent parameters are the log levels for writing to the file. In ErrorLevel mode, only error - level logs are recorded.
		// errorLogCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), highPriority) // The third and subsequent parameters are the log levels for writing to the file. In ErrorLevel mode, only error - level logs are recorded.

		// coreArr = append(coreArr, infoLogCore, errorLogCore)
	} else {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdin)), zapcore.InfoLevel) // The third and subsequent parameters are the log levels for writing to the file. In ErrorLevel mode, only error - level logs are recorded.

		coreArr = append(coreArr, consoleCore)
	}

	log := zap.New(zapcore.NewTee(coreArr...), zap.AddCaller()) // zap.AddCaller() is used to display the file name and line number and can be omitted.
	// defer log.Sync()

	return &obLogger{
		L: log,
		S: log.Sugar(),
	}
}

func (l *obLogger) GetLogger() *zap.Logger               { return l.L }
func (l *obLogger) GetSugaredLogger() *zap.SugaredLogger { return l.S }

func (l *obLogger) Info(ctx context.Context, message string, params ...interface{}) {
	requestId := ctx.Value("requestId").(string)
	if len(params) > 0 {
		l.L.Info(message, zap.String("requestId", requestId), zap.Any("params: ", params))
		return
	}
	l.L.Info(message, zap.String("requestId", requestId))
}

func (l *obLogger) Error(ctx context.Context, message string, err interface{}) {
	requestId := ctx.Value("requestId").(string)
	if errMsg, ok := err.(string); ok {
		l.L.Error(message, zap.String("requestId", requestId), zap.String("err", errMsg))
	} else if errObj, ok := err.(error); ok {
		l.L.Error(message, zap.String("requestId", requestId), zap.String("err:", errObj.Error()))
	}
}

func (l *obLogger) Errors(ctx context.Context, message string, err ...interface{}) {
	requestId := ctx.Value("requestId").(string)
	if len(err) > 0 {
		l.L.Info(message, zap.String("requestId", requestId), zap.Any("errors: ", err))
		return
	}
	l.L.Info(message, zap.String("requestId", requestId))
}

func (l *obLogger) Warn(ctx context.Context, message string, params ...interface{}) {
	requestId := ctx.Value("requestId").(string)
	if len(params) > 0 {
		l.L.Warn(message, zap.String("requestId", requestId), zap.Any("params: ", params))
		return
	}
	l.L.Warn(message, zap.String("requestId", requestId))
}

func (l *obLogger) InfoWithMask(ctx context.Context, message, secret string) {
	requestId := ctx.Value("requestId").(string)
	len := len(secret)
	if len > 0 && len <= 8 {
		// print all mask if secret is too short
		secret = "*************"
	} else if len > 8 {
		// Only print last 4 characters
		secret = "*************" + secret[len-4:]
	}
	l.L.Info(message, zap.String("requestId", requestId), zap.String("data", secret))
}
