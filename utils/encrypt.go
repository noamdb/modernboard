package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/spf13/viper"
)

func EncryptString(s string) string {
	if s == "" {
		return ""
	}
	h := sha256.New()
	io.WriteString(h, viper.GetString("default_salt")+s)
	e := hex.EncodeToString(h.Sum(nil))
	return e[:4] + e[len(e)-4:len(e)]
}
