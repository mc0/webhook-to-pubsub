package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

func addPadding(blockSize int, d []byte) []byte {
	padding := make([]byte, blockSize-len(d)%blockSize)
	return append(d, padding...)
}

func encrypt(block cipher.Block, text []byte) ([]byte, error) {
	bs := block.BlockSize()
	text = addPadding(bs, text)
	res := make([]byte, bs+len(text))
	iv := res[:bs]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	m := cipher.NewCBCEncrypter(block, iv)
	m.CryptBlocks(text, text)
	copy(res[bs:], text)

	return res, nil
}

func decrypt(block cipher.Block, text []byte) ([]byte, error) {
	bs := block.BlockSize()
	if len(text) < bs {
		return nil, errors.New("cipher text too short")
	}

	iv := text[:bs]
	text = text[bs:]
	m := cipher.NewCBCDecrypter(block, iv)
	m.CryptBlocks(text, text)
	text = bytes.TrimRight(text, "\x00")

	return text, nil
}

func decryptURLBase64(block cipher.Block, input string) ([]byte, error) {
	text, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}

	text, err = decrypt(block, text)
	return text, err
}
