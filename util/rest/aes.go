package rest

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type AESDecoder struct {
	cipherBlock cipher.Block
}

func NewAESDecoder(key []byte) (*AESDecoder, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESDecoder{
		cipherBlock: c,
	}, nil
}

func (d *AESDecoder) DecodeBody(request *http.Request) (io.Reader, error) {
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("message is empty")
	}
	if len(data) <= aes.BlockSize {
		return nil, errors.New("encrypted message is too short")
	}

	iv := data[:aes.BlockSize]
	msg := data[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(d.cipherBlock, iv)
	cfb.XORKeyStream(msg, msg)
	fmt.Println(string(msg))
	return bytes.NewReader(msg), nil

}
