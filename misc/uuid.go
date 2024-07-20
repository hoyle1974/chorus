package misc

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
)

func encode(data []byte) string {
	encoded := base64.RawStdEncoding.EncodeToString(data)
	encoded = string(bytes.ReplaceAll([]byte(encoded), []byte("/"), []byte("_")))
	encoded = string(bytes.ReplaceAll([]byte(encoded), []byte("+"), []byte("-")))
	encoded = string(bytes.ReplaceAll([]byte(encoded), []byte("="), []byte(".")))
	return encoded
}

func UUIDString() string {
	token := make([]byte, 6)
	rand.Read(token)
	encodedString := encode(token)

	return encodedString
}
