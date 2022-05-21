package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	PwSaltBytes = 16
	ApiKeyBytes = 16
)

// Encrypt https://gist.github.com/mickelsonm/e1bf365a149f3fe59119
func encrypt(key []byte, message string) (string, error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//returns to base64 encoded string
	encryptedMessage := base64.StdEncoding.EncodeToString(cipherText)
	return encryptedMessage, nil
}

// Decrypt https://gist.github.com/mickelsonm/e1bf365a149f3fe59119
func Decrypt(key []byte, secureMessage string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(secureMessage)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short! ")
		return "", err
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decodedMessage := string(cipherText)
	return decodedMessage, nil
}

func scryptPassword(password string, salt []byte) (string, error) {
	encryptedKey, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedKey), nil
}

func createSalt() ([]byte, error) {
	salt := make([]byte, PwSaltBytes)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func generateApiKeySecret() (string, error) {
	randomBytes := make([]byte, ApiKeyBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", errors.New("Key generation failed ")
	}
	clearTextApiKey := base64.StdEncoding.EncodeToString(randomBytes)
	return clearTextApiKey, nil
}
