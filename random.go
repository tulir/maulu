// From https://stackoverflow.com/a/31832326/2120293
package main

import (
	"math/rand"
	"strings"
	"time"
)

const allowedCharacters = "abcdefghijklmnopqrstuvwxyz123456789-_"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

var src = rand.NewSource(time.Now().UnixNano())

func randomShortURL() string {
	var sb strings.Builder
	sb.Grow(6)
	for i, cache, remain := 4, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(allowedCharacters) {
			sb.WriteByte(allowedCharacters[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return sb.String()
}
