package main

import (
	"os"

	"github.com/0xFilosoF/pow-ddos-guard/internal/client"
	"github.com/0xFilosoF/pow-ddos-guard/pkg/config"
	gracefullserver "github.com/0xFilosoF/pow-ddos-guard/pkg/gracefullServer"
	"github.com/0xFilosoF/pow-ddos-guard/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	clientConfig := config.New[config.ClientParams]("configs/client.yaml")

	var logPreset string
	if clientConfig.Debug {
		logPreset = "development"
	} else {
		logPreset = "production"
	}
	logLevel, logErr := zapcore.ParseLevel(clientConfig.Log.Level)
	if logErr != nil {
		panic(logErr)
	}

	sync := logger.Init(logPreset, clientConfig.App.Name, logLevel)

	cli := client.New(clientConfig)

	if err := gracefullserver.New(cli).Start(); err != nil {
		zap.L().Error("TCP client exited with error", zap.Error(err))
		_ = sync()
		os.Exit(1)
	}
	_ = sync()
}
