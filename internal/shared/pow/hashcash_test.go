package pow_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/pow"
)

//nolint:gochecknoglobals // for tests
var stampTests = []struct {
	bits      uint
	saltLen   uint
	extension string
	resource  string
}{
	{20, 8, "", "abc"},
	{10, 10, "asdf", "something"},
	{20, 10, "abc", "something"},
	{15, 4, "", "someone@example.net"},
}

func TestStampFormat(t *testing.T) {
	expectedDate := time.Now().UTC().Format("0601021504")

	for _, tt := range stampTests {
		h := pow.New(tt.bits, tt.saltLen, 10, tt.extension)

		stamp, err := h.Mint(tt.resource)
		if err != nil {
			t.Errorf("Mint failed for %s with error %v", tt.resource, err)
		}
		fields := strings.Split(stamp, ":")
		if len(fields) != 7 {
			t.Errorf("Expected 7 fields got %d", len(fields))
		}
		ver, err := strconv.Atoi(fields[0])
		if err != nil {
			t.Errorf("Expected version 1, got error %v", err)
		}
		if ver != 1 {
			t.Errorf("Expected version 1, got %d", ver)
		}
		bits, err := strconv.ParseUint(fields[1], 10, 32)
		if err != nil {
			t.Errorf("Expected %d bits, got error %v", tt.bits, err)
		}
		if uint(bits) != tt.bits {
			t.Errorf("Expected %d bits, got %d", tt.bits, bits)
		}
		date := fields[2]
		then, _ := time.ParseInLocation("060102150405", date, time.UTC)
		if then.Format("0601021504") != expectedDate {
			t.Errorf("Expected %s date, got %s", expectedDate, then)
		}
		resource := fields[3]
		if resource != tt.resource {
			t.Errorf("Expected %s resource, got %s", tt.resource, resource)
		}
		extension := fields[4]
		if extension != tt.extension {
			t.Errorf("Expected %s extra, got %s", tt.extension, extension)
		}
		salt := fields[5]
		if uint(len(salt)) != tt.saltLen {
			t.Errorf("Expected %d salt chars, got %d", tt.saltLen, len(salt))
		}
		counter := fields[6]
		if counter == "" {
			t.Errorf("Counter field is empty")
		}
	}
}

//nolint:gochecknoglobals // for tests
var checkNoDateTests = []string{
	"1:20:260115194732:abc::pCRllXP-:e18e51a1e7da96c4",
	"1:20:260115194732:something::KjVqVwtp:d10e8f380953a2b8",
	"1:20:260115194732:someone@example.net::uQ4WPxL3:499ae11e04f9f7ad"}

func TestCheckNoDate(t *testing.T) {
	h := pow.Default()
	for _, stamp := range checkNoDateTests {
		if !h.CheckNoDate(stamp) {
			t.Errorf("Failed for %s", stamp)
		}
	}
}

//nolint:gochecknoglobals // for tests
var mintAndCheckTests = []string{
	"abc",
	"something",
	"someone@example.net"}

func TestMintAndCheck(t *testing.T) {
	h := pow.Default()
	for _, resource := range mintAndCheckTests {
		stamp, err := h.Mint(resource)
		if err != nil {
			t.Errorf("Mint failed for %s with error %v", resource, err)
		}
		if !h.Check(stamp) {
			t.Errorf("Check failed for %s", resource)
		}
	}
}
