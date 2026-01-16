package handler

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"time"

	"github.com/0xFilosoF/pow-ddos-guard/internal/server/quote"
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
	defer rpcConn.Close()

	// notify challenge
	if rpcErr := rpcConn.Notify(ctx, "wow.challenge", ch); rpcErr != nil {
		zap.L().Error("Failed to notify challenge", zap.Error(rpcErr))
		return
	}

	<-rpcConn.DisconnectNotify()
}

func (sh *sessionHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	switch req.Method {
	case "wow.verify":
		if req.Params == nil {
			_ = conn.ReplyWithError(
				ctx,
				req.ID,
				&jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: "missing params"},
			)
			return
		}

		var vr challenge.VerifyRequest
		if err := json.Unmarshal(*req.Params, &vr); err != nil {
			_ = conn.ReplyWithError(
				ctx,
				req.ID,
				&jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: err.Error()},
			)
			return
		}

		if err := sh.ch.Verify(vr); err != nil {
			_ = conn.ReplyWithError(
				ctx,
				req.ID,
				&jsonrpc2.Error{Code: jsonrpc2.CodeInvalidRequest, Message: err.Error()},
			)
			return
		}

		_ = conn.Reply(ctx, req.ID, challenge.VerifyResponse{Quote: quote.Random()})
		return
	default:
		_ = conn.ReplyWithError(
			ctx,
			req.ID,
			&jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: "method not found"},
		)
		return
	}
}
