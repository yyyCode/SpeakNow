package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level, output string) (*zap.Logger, error) {
	lvl := zap.NewAtomicLevel()
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl.SetLevel(zap.InfoLevel)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var ws zapcore.WriteSyncer
	if output == "stdout" || output == "" {
		ws = zapcore.AddSync(os.Stdout)
	} else {
		f, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		ws = zapcore.AddSync(f)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		ws,
		lvl,
	)
	return zap.New(core, zap.AddCaller()), nil
}
