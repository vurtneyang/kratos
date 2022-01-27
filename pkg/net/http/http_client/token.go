package http_client

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	_envSignSecret = "SIGN_SECRET"
)

func GenSign() (string, error) {
	secret := os.Getenv(_envSignSecret)
	if secret == "" {
		return "", errors.New("environment variable SIGN_SECRET is not set")
	}

	timestamp := time.Now().Unix()
	sign := GetSign(secret, timestamp)
	return sign, nil
}

func GetSign(secret string, timestamp int64) string {
	bytes := md5.Sum([]byte(fmt.Sprintf("%d,%s", timestamp, secret)))
	return fmt.Sprintf("%s,%d", hex.EncodeToString(bytes[:]), timestamp)
}
