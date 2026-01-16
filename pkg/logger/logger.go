package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(preset, srvName string, logLevel zapcore.Level) func() error {
	var cfg zap.Config
	switch preset {
	case "production":
		cfg = zap.NewProductionConfig()
		cfg.Sampling = nil
	case "development":
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.LineEnding = ""
	default:
		panic(fmt.Sprintf("Invalid preset %s", preset))
	}

	cfg.InitialFields = map[string]any{"service": srvName}
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
		pae.AppendInt64(t.UnixMilli())
	}
	cfg.Level.SetLevel(logLevel)

	log := zap.Must(cfg.Build())
	zap.ReplaceGlobals(log)
	return log.Sync
}
