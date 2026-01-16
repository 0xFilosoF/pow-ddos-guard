package pow

const (
	bitLen       = 8
	maxByteValue = 0xFF
)

// hasLeadingZeroBits classis anti-ddos solution.
func hasLeadingZeroBits(b []byte, bits int) bool {
	if bits <= 0 {
		return true
	}

	full := bits / bitLen
	for i := range full {
		if b[i] != 0 {
			return false
		}
	}

	rem := bits % bitLen
	if rem == 0 {
		return true
	}

	mask := byte(maxByteValue) << (bitLen - rem)
	return (b[full] & mask) == 0
}
