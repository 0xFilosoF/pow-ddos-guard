package challenge

import (
	"errors"
)

type VerifyRequest struct {
	ID    string `json:"id"`
	Nonce string `json:"nonce"` // hashcash v1 stamp
}

type VerifyResponse struct {
	Quote string `json:"quote"`
}

func (ch *Challenge) Verify(req VerifyRequest) error {
	if req.ID != ch.ID {
		return errors.New("challenge id mismatch")
	}
	if ok := ch.hc.Check(req.Nonce); !ok {
		return errors.New("challenge not valid or expired")
	}
	return nil
}
