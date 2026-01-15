package server

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/0xFilosoF/pow-ddos-guard/internal/server/handler"
	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/pow"
	"github.com/0xFilosoF/pow-ddos-guard/pkg/config"
	"go.uber.org/zap"
)

type Server struct {
	Hashcash *pow.Hashcash
	cfg      *config.Config[config.ServerParams]
}

func New(cfg *config.Config[config.ServerParams]) *Server {
	difficulty, saltLen, ttl := cfg.Params.PoW.Difficulty, cfg.Params.PoW.SaltLen, cfg.Params.PoW.ChallengeTTL
	hc := pow.New(difficulty, saltLen, ttl, "")

	return &Server{
		Hashcash: hc,
		cfg:      cfg,
	}
}

func (s *Server) Start() error {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", s.cfg.App.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	if s.cfg.TLS.Enabled {
		cert, tlsErr := tls.LoadX509KeyPair(s.cfg.TLS.Cert, s.cfg.TLS.Key)
		if tlsErr != nil {
			return tlsErr
		}

		ln = tls.NewListener(ln, &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS13,
		})
	}

	zap.L().Info("Listening TCP server started", zap.String("address", s.cfg.App.Addr))

	for {
		rawConn, lnErr := ln.Accept()

		if lnErr != nil {
			zap.L().Error("Failed to accept connection", zap.Error(lnErr))
			continue
		}

		remoteAddr := rawConn.RemoteAddr().String()
		zap.L().Info("New client connected", zap.String("address", remoteAddr))

		go handler.NewConn(s.cfg, s.Hashcash, rawConn)
	}
}

// func (s *Server) Shutdown(ctx context.Context) error {
// 	errCh := make(chan error, 1)
//
// 	wg := sync.WaitGroup{}
// 	const workersCount = 2
// 	wg.Add(workersCount)
//
// 	go func() {
// 		defer wg.Done()
// 		s.grpcSrv.GracefulStop()
// 	}()
//
// 	go func() {
// 		defer wg.Done()
// 		if err := s.collector.Shutdown(ctx); err != nil {
// 			errCh <- err
// 		}
// 	}()
//
// 	wg.Wait()
// 	close(errCh)
//
// 	if err := <-errCh; err != nil {
// 		return err
// 	}
//
// 	return nil
// }
