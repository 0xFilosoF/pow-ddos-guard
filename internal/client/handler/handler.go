package handler

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/challenge"
	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/pow"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type ClientHandler struct {
	hc          *pow.Hashcash
	challengeCh chan handlerData
}

type handlerData struct {
	rpcConn *jsonrpc2.Conn
	ch      challenge.Challenge
	reqCtx  context.Context
}

func New(ctx context.Context, hc *pow.Hashcash) *ClientHandler {
	const chCap = 1
	h := &ClientHandler{
		hc:          hc,
		challengeCh: make(chan handlerData, chCap),
	}

	go h.run(ctx)

	return h
}

func (h *ClientHandler) GenerateRequestsAsync(ctx context.Context, rpcConn *jsonrpc2.Conn) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := rpcConn.Notify(ctx, "wow.create", nil); err != nil {
				zap.L().Error("Failed to create challenge", zap.Error(err))
			}
		}
	}
}

func (h *ClientHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	zap.L().Info(
		"Received request",
		zap.String("id", req.ID.String()),
		zap.String("method", req.Method),
	)

	switch req.Method {
	case "wow.challenge":
		var ch challenge.Challenge
		if req.Params == nil {
			zap.L().Error("missing params")
			return
		}
		if err := json.Unmarshal(*req.Params, &ch); err != nil {
			zap.L().Error("Got bad challenge", zap.Error(err))
			return
		}

		select {
		case h.challengeCh <- handlerData{rpcConn: conn, ch: ch, reqCtx: ctx}:
		default:
			zap.L().Warn("Challenge channel is full")
		}
		return
	default:
	}
}

func (h *ClientHandler) run(ctx context.Context) {
	defer close(h.challengeCh)

	for {
		select {
		case <-ctx.Done():
			return
		case challengeCh := <-h.challengeCh:
			rpcConn, ch := challengeCh.rpcConn, challengeCh.ch

			expiredAt := time.Unix(ch.ExpiresAt, 0)
			h.hc.Update(ch.Difficulty, time.Until(expiredAt))

			nonce, err := h.hc.Mint(ch.ChallengeB64)
			if err != nil {
				zap.L().Error("Got bad challenge", zap.Error(err))
				continue
			}
			if ok := h.hc.Check(nonce); !ok {
				zap.L().
					Error("Bad challenge after local checking",
						zap.Error(errors.New("challenge not valid or expired")))
				continue
			}

			vr := &challenge.VerifyRequest{
				ID:    ch.ID,
				Nonce: nonce,
			}

			var resp challenge.VerifyResponse
			if err = rpcConn.Call(ctx, "wow.verify", vr, &resp); err != nil {
				zap.L().Error("Verify call error", zap.Error(err))
				continue
			}

			zap.L().Info("Got it", zap.String("quote", resp.Quote))
		}
	}
}
