package rest

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

type AESDecoderEncoder struct {
	cipherBlock cipher.Block
}

func NewAESDecoderEncoder(key []byte) (*AESDecoderEncoder, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESDecoderEncoder{
		cipherBlock: c,
	}, nil
}

func (d *AESDecoderEncoder) DecodeBody(request *http.Request) (io.Reader, error) {
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
	return bytes.NewReader(msg), nil
}

func (d *AESDecoderEncoder) Encode(rw http.ResponseWriter) (http.ResponseWriter, error) {
	return &AESResponseWriter{rw, d.cipherBlock}, nil
}

type AESResponseWriter struct {
	http.ResponseWriter
	cipherBlock cipher.Block
}

func (a *AESResponseWriter) Write(msg []byte) (int, error) {
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	cfb := cipher.NewCFBEncrypter(a.cipherBlock, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(msg))

	_, err := a.ResponseWriter.Write(ciphertext)
	if err != nil {
		return 0, err
	} else {
		return len(msg), err
	}
}
