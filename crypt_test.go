package main

import (
	"crypto/aes"
	"testing"
)

func TestCrypt(t *testing.T) {
	key := "aaaaaaaaaaaaaaaa"
	var err error
	block, err := aes.NewCipher([]byte(key))
	if nil != err {
		t.Fatal("AES block failed", err)
	}

	result, err := encrypt(block, []byte("testing"))
	if nil != err || len(result) == 0 {
		t.Fatal("Encrypt failed", err)
	}
}
