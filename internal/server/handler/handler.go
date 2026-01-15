package handler

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/challenge"
	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/pow"
	"github.com/0xFilosoF/pow-ddos-guard/pkg/config"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type sessionHandler struct {
	ch *challenge.Challenge
	// mu sync.Mutex // if conn not closed
}

func NewConn(cfg *config.Config[config.ServerParams], hc *pow.Hashcash, rawConn net.Conn) {
	defer rawConn.Close()
	ctx := context.Background()

	if cfg.TLS.Enabled {
		if tc, ok := rawConn.(*tls.Conn); ok {
			_ = tc.SetDeadline(time.Now().UTC().Add(cfg.Params.Handshake))
			if err := tc.HandshakeContext(ctx); err != nil {
				zap.L().Error("TLS handshake error", zap.Error(err))
				return
			}
			_ = tc.SetDeadline(time.Time{})
		}
	}

	const multiplier = 2
	_, ttl := hc.GetChallenge()
	_ = rawConn.SetReadDeadline(time.Now().UTC().Add(ttl * multiplier))
	_ = rawConn.SetWriteDeadline(time.Now().UTC().Add(ttl * multiplier))

	ch, err := challenge.New(hc)
	if err != nil {
		zap.L().Error("Failed to create a challenge", zap.Error(err))
		return
	}

	handler := &sessionHandler{
		ch: ch,
	}

	stream := jsonrpc2.NewPlainObjectStream(rawConn)
	rpcConn := jsonrpc2.NewConn(ctx, stream, handler)

	// notify challenge
	if rpcErr := rpcConn.Notify(ctx, "wow.challenge", ch); rpcErr != nil {
		zap.L().Error("Failed to notify challenge", zap.Error(rpcErr))
		_ = rpcConn.Close()
		return
	}

	<-rpcConn.DisconnectNotify()
}

func (sh *sessionHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	// TODO:
}
