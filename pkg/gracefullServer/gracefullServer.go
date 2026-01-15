package gracefullserver

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type GracefullServer interface {
	Start() error
	Shutdown(ctx context.Context) error
}

type Server struct {
	s GracefullServer
}

func New(s GracefullServer) *Server {
	return &Server{s: s}
}

func (gs *Server) Start() error {
	errCh := make(chan error, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := gs.s.Start(); err != nil && !errors.Is(err, net.ErrClosed) {
			errCh <- err
		}
	}()

	var err error

	select {
	case sig := <-sigCh:
		zap.L().Info("received shutdown signal", zap.Any("signal", sig))
	case err = <-errCh:
		zap.L().Error("server stopped with error", zap.Error(err))
	}

	if err == nil {
		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if shutdownErr := gs.s.Shutdown(ctx); shutdownErr != nil {
			zap.L().Error("graceful shutdown failed", zap.Error(shutdownErr))
			return shutdownErr
		}

		zap.L().Info("server gracefully shutdown")

		return nil
	}

	return err
}
