package handler

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/0xFilosoF/pow-ddos-guard/internal/server/quote"
	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/challenge"
	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/pow"
	"github.com/0xFilosoF/pow-ddos-guard/pkg/config"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type sessionHandler struct {
	hashcash *pow.Hashcash
	ch       *challenge.Challenge
	rawConn  net.Conn
	mu       sync.Mutex
}

func NewConn(cfg *config.Config[config.ServerParams], hc *pow.Hashcash, rawConn net.Conn) {
	defer rawConn.Close()
	ctx := context.Background()

	if cfg.TLS.Enabled {
		if tc, ok := rawConn.(*tls.Conn); ok {
			_ = tc.SetDeadline(time.Now().Add(cfg.Params.Handshake))
			if err := tc.HandshakeContext(ctx); err != nil {
				zap.L().Error("TLS handshake error", zap.Error(err))
				return
			}
			_ = tc.SetDeadline(time.Time{})
		}
	}

	const multiplier = 2
	_, ttl := hc.GetChallenge()
	_ = rawConn.SetReadDeadline(time.Now().Add(ttl * multiplier))
	_ = rawConn.SetWriteDeadline(time.Now().Add(ttl * multiplier))

	handler := &sessionHandler{
		hashcash: hc,
		rawConn:  rawConn,
	}
	if err := handler.generateChallenge(); err != nil {
		zap.L().Error("Failed to generate a challenge", zap.Error(err))
		return
	}

	stream := jsonrpc2.NewPlainObjectStream(rawConn)
	rpcConn := jsonrpc2.NewConn(ctx, stream, handler)
	defer rpcConn.Close()

	// notify challenge
	if rpcErr := rpcConn.Notify(ctx, "wow.challenge", handler.ch); rpcErr != nil {
		zap.L().Error("Failed to notify challenge", zap.Error(rpcErr))
		return
	}

	<-rpcConn.DisconnectNotify()
}

func (sh *sessionHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	zap.L().Info(
		"Received request",
		zap.String("address", sh.rawConn.RemoteAddr().String()),
		zap.String("id", req.ID.String()),
		zap.String("method", req.Method),
	)

	switch req.Method {
	case "wow.create":
		if sh.ch == nil {
			if err := sh.generateChallenge(); err != nil {
				_ = conn.ReplyWithError(
					ctx,
					req.ID,
					&jsonrpc2.Error{
						Code:    jsonrpc2.CodeInternalError,
						Message: "failed to generate a challenge",
					},
				)
				return
			}
		}

		if err := conn.Notify(ctx, "wow.challenge", sh.ch); err != nil {
			zap.L().Error("Failed to notify challenge", zap.Error(err))
			return
		}

		return
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

		if sh.ch == nil {
			_ = conn.ReplyWithError(
				ctx,
				req.ID,
				&jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidRequest,
					Message: "challenge not found, try again",
				},
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

		// NOTE: generate new challenge with new ID is trustful
		// and does not allow replay attack with mutex
		if err := sh.generateChallenge(); err != nil {
			zap.L().Error("Failed to generate a challenge", zap.Error(err))
			sh.ch = nil
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

func (sh *sessionHandler) generateChallenge() error {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	ch, err := challenge.New(sh.hashcash)
	if err != nil {
		return err
	}

	sh.ch = ch
	return nil
}
