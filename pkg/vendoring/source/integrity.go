package source

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// IntegrityChecksum translates registry SRI into go-getter's checksum syntax,
// preferring the strongest supported algorithm rather than token order.
func IntegrityChecksum(value string) (string, error) {
	for _, algorithm := range []struct {
		name string
		size int
	}{{"sha512", 64}, {"sha256", 32}} {
		for _, token := range strings.Fields(value) {
			prefix := algorithm.name + "-"
			if !strings.HasPrefix(token, prefix) {
				continue
			}
			data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(token, prefix))
			if err != nil || len(data) != algorithm.size {
				return "", fmt.Errorf("invalid npm integrity value")
			}
			return algorithm.name + ":" + hex.EncodeToString(data), nil
		}
	}
	return "", fmt.Errorf("npm integrity must include SHA-256 or SHA-512")
}
