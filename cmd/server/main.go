package main

import (
	"github.com/0xFilosoF/pow-ddos-guard/internal/server"
	"github.com/0xFilosoF/pow-ddos-guard/pkg/config"
	gracefullserver "github.com/0xFilosoF/pow-ddos-guard/pkg/gracefullServer"
	"github.com/0xFilosoF/pow-ddos-guard/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	serverConfig := config.New[config.ServerParams]("configs/server.yaml")

	var logPreset string
	if serverConfig.Debug {
		logPreset = "development"
	} else {
		logPreset = "production"
	}
	logLevel, logErr := zapcore.ParseLevel(serverConfig.Log.Level)
	if logErr != nil {
		panic(logErr)
	}

	sync := logger.Init(logPreset, serverConfig.App.Name, logLevel)
	//nolint:errcheck // defer sync zap logger
	defer sync()

	tcpServer := server.New(serverConfig)

	if err := gracefullserver.New(tcpServer).Start(); err != nil {
		zap.L().Error("TCP server exited with error", zap.Error(err))
		// NOTE: not using os.Exit because we need call gracefull shutdown
		panic(err)
	}
}
