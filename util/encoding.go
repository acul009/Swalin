package util

import "encoding/base64"

func Base64Encode(data []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(data))
}

func Base64Decode(data []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(data))
}
