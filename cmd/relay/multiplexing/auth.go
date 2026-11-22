package multiplexing

import "crypto/subtle"

func Authorize(key string, secretKey string) bool {
	return subtle.ConstantTimeCompare([]byte(key), []byte(secretKey)) == 1
}
