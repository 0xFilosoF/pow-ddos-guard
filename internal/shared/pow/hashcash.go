package pow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	maxIterations    int    = 1 << 32        // Max iterations to find a solution
	hashcashV1Length int    = 7              // hashcash stamp v1 format (v:bits:date:resource:extension:nonce)
	timeFormat       string = "060102150405" // YYMMDDhhmmss
)

var (
	ErrInvalidStamp = errors.New("invalid stamp")
	ErrExpired      = errors.New("stamp expired")
)

// Hashcash provides an implementation of hashcash v1 (anti-ddos friendly)
// Stamp format: ver:bits:date:resource:extension:salt:nonce.
type Hashcash struct {
	bits      uint          // Number of leading zero bits required (difficulty)
	expired   time.Duration // TTL window in seconds (0 => ignore date)
	saltLen   uint          // Random salt length (rand field length in chars)
	extension string        // Extension field
	now       func() time.Time
}

// New creates a new Hash with specified options.
// bits: leading zero bits required
// saltLen: length of rand field (chars) - base64 chars are fine
// extension: extension field
// expired: if>0, stamp date must be within this TTL window.
func New(bits uint, saltLen uint, expired time.Duration, extension string) *Hashcash {
	return &Hashcash{
		bits:      bits,
		expired:   expired,
		saltLen:   saltLen,
		extension: extension,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func Default() *Hashcash {
	const (
		defaultBits    uint          = 20
		defaultSaltLen uint          = 8
		defaultExpired time.Duration = 30 * time.Second
	)
	return New(defaultBits, defaultSaltLen, defaultExpired, "")
}

func (hc *Hashcash) GetChallenge() (uint, time.Duration) {
	return hc.bits, hc.expired
}

func (hc *Hashcash) Update(newBits uint, newExpired time.Duration) {
	hc.bits = newBits
	hc.expired = newExpired
}

// Mint a new hashcash v1 stamp for resource.
// It searches for a nonce such that sha256(stamp) has `bits` leading zero bits.
func (hc *Hashcash) Mint(resource string) (string, error) {
	salt, err := hc.getSalt()
	if err != nil {
		return "", err
	}

	// seconds-precision timestamp
	date := hc.now().Format(timeFormat)

	var start uint64
	if err = binary.Read(rand.Reader, binary.BigEndian, &start); err != nil {
		return "", err
	}

	// NOTE: we can use some crptography solution for nonce
	for i := range maxIterations {
		//nolint:gosec // overflow not possible with maxIterations
		nonce := start + uint64(i)

		stamp := fmt.Sprintf("1:%d:%s:%s:%s:%s:%x",
			hc.bits, date, resource, hc.extension, salt, nonce)

		if hc.checkZerosBits(stamp) {
			return stamp, nil
		}
	}

	return "", errors.New("no solution within maxIterations")
}

// Check validates stamp including date (if expired > 0).
func (hc *Hashcash) Check(stamp string) bool {
	p, err := hc.validate(stamp)
	if err != nil {
		return false
	}
	if hc.expired != 0 {
		if err = hc.checkDate(p); err != nil {
			return false
		}
	}
	return hc.checkZerosBits(stamp)
}

// CheckNoDate validates stamp ignoring date.
func (hc *Hashcash) CheckNoDate(stamp string) bool {
	_, err := hc.validate(stamp)
	if err != nil {
		return false
	}
	return hc.checkZerosBits(stamp)
}

func (hc *Hashcash) validate(stamp string) (*ParsedV1, error) {
	p, err := Parse(stamp)
	if err != nil {
		return nil, err
	}
	if p.Bits < hc.bits {
		return nil, ErrInvalidStamp
	}
	return p, nil
}

func (hc *Hashcash) getSalt() (string, error) {
	if hc.saltLen == 0 {
		hc.saltLen = 8
	}
	buf := make([]byte, hc.saltLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	// RawURLEncoding without '+', '/', '=', etc
	salt := base64.RawURLEncoding.EncodeToString(buf)
	if uint(len(salt)) >= hc.saltLen {
		return salt[:hc.saltLen], nil
	}

	return salt, nil
}

func (hc *Hashcash) checkZerosBits(stamp string) bool {
	sum := sha256.Sum256([]byte(stamp))
	//nolint:gosec // overflow not possible, bits is quite small
	return hasLeadingZeroBits(sum[:], int(hc.bits))
}

func (hc *Hashcash) checkDate(p *ParsedV1) error {
	now := hc.now()

	if p.Date.After(now.Add(time.Second)) {
		return ErrExpired
	}
	if now.Sub(p.Date) > hc.expired {
		return ErrExpired
	}

	return nil
}

// ParsedV1 gives you parsed fields if you want to bind resource/salt on service side.
type ParsedV1 struct {
	Version   string
	Bits      uint
	Date      time.Time
	Resource  string
	Extension string
	Salt      string
	Nonce     string
}

// Parse parses a v1 stamp (does not verify PoW).
func Parse(stamp string) (*ParsedV1, error) {
	fields := strings.Split(stamp, ":")
	if len(fields) != hashcashV1Length {
		return nil, ErrInvalidStamp
	}
	if fields[0] != "1" {
		return nil, errors.New("invalid version")
	}
	bits64, err := strconv.ParseUint(fields[1], 10, 32)
	if err != nil {
		return nil, ErrInvalidStamp
	}
	dt, err := time.ParseInLocation(timeFormat, fields[2], time.UTC)
	if err != nil {
		return nil, ErrInvalidStamp
	}

	if _, err = strconv.ParseUint(fields[6], 16, 64); err != nil {
		return nil, ErrInvalidStamp
	}
	return &ParsedV1{
		Version:   fields[0],
		Bits:      uint(bits64),
		Date:      dt,
		Resource:  fields[3],
		Extension: fields[4],
		Salt:      fields[5],
		Nonce:     fields[6],
	}, nil
}
