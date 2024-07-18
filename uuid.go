package main

import (
	"crypto/rand"
	"encoding/base64"
)

func UUIDString() string {
	token := make([]byte, 6)
	rand.Read(token)

	return base64.StdEncoding.EncodeToString(token)
}
