package challenge

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/pow"
)

type Challenge struct {
	ID           string `json:"id"`
	ChallengeB64 string `json:"challenge"`  // base64(random bytes)
	Difficulty   uint   `json:"difficulty"` // leading-zero bits
	ExpiresAt    int64  `json:"expires_at"` // unix seconds
	hc           *pow.Hashcash
}

func New(hc *pow.Hashcash) (*Challenge, error) {
	const rawLen = 32
	b := make([]byte, rawLen)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	chB64 := base64.RawURLEncoding.EncodeToString(b)
	id := chB64[:12]

	difficulty, ttl := hc.GetChallenge()

	return &Challenge{
		ID:           id,
		ChallengeB64: chB64,
		Difficulty:   difficulty,
		ExpiresAt:    time.Now().UTC().Add(ttl).Unix(),
		hc:           hc,
	}, nil
}
