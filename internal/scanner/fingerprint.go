package scanner

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenerateFingerprint(parts ...string) string {
	hasher := sha256.New()

	for _, part := range parts {
		hasher.Write([]byte(part))
		hasher.Write([]byte("|"))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}
