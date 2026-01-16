package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/challenge"
	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/pow"
	"github.com/sourcegraph/jsonrpc2"
	"go.uber.org/zap"
)

type ClientHandler struct {
	conn *jsonrpc2.Conn
	hc   *pow.Hashcash
}

func (h *ClientHandler) Update(conn *jsonrpc2.Conn, hc *pow.Hashcash) {
	h.conn = conn
	h.hc = hc
}

func (h *ClientHandler) Handle(_ context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	switch req.Method {
	case "wow.challenge":
		defer h.conn.Close()

		var ch challenge.Challenge
		if req.Params == nil {
			zap.L().Error("missing params")
			return
		}
		if err := json.Unmarshal(*req.Params, &ch); err != nil {
			zap.L().Error("Got bad challenge", zap.Error(err))
			return
		}

		h.hc.Bits = ch.Difficulty
		h.hc.Expired = time.Duration(ch.ExpiresAt)
		nonce, err := h.hc.Mint(ch.ChallengeB64)
		if err != nil {
			zap.L().Error("Got bad challenge", zap.Error(err))
			return
		}
		// if ok := h.hc.Check(nonce); !ok {
		// 	zap.L().
		// 		Error("Got bad challenge", zap.Error(errors.New("challenge not valid or expired")))
		// 	return
		// }

		vr := &challenge.VerifyRequest{
			ID:    ch.ID,
			Nonce: nonce,
		}
		fmt.Println(vr)

		var resp challenge.VerifyResponse
		if err = conn.Call(context.Background(), "wow.verify", vr, &resp); err != nil {
			zap.L().Error("Verify call error", zap.Error(err))
			return
		}

		zap.L().Info("Got it", zap.String("quote", resp.Quote))
		return
	default:
	}
}
