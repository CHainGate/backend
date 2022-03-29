package utils

import (
	"encoding/hex"
	"testing"
)

type AesTest struct {
	key              []byte
	clearTextMessage string
	encryptedMessage string
}

const (
	saltLength   = 16
	apiKeyLength = 24 // initially 16 random bytes after base64 encoding 24 chars
)

var aesTest = AesTest{
	key:              []byte("secret_123456789"),
	clearTextMessage: "my clear text message",
}

// Do not parallelize this test! Unless there is a way to mock aes to generate the same encryption text
func TestEncrypt(t *testing.T) {
	encryptedMessage, err := Encrypt(aesTest.key, aesTest.clearTextMessage)
	if err != nil {
		t.Errorf("Message encryption failed. Error: %s", err.Error())
	}
	aesTest.encryptedMessage = encryptedMessage
}

// Do not parallelize this test! Unless there is a way to mock aes to generate the same encryption text
func TestDecrypt(t *testing.T) {
	decryptedMessage, err := Decrypt(aesTest.key, aesTest.encryptedMessage)
	if err != nil {
		t.Fatalf("Message decryption failed. Error: %s", err.Error())
	}
	if decryptedMessage != aesTest.clearTextMessage {
		t.Errorf("Expected message to be %s, but got %s", aesTest.clearTextMessage, decryptedMessage)
	}
}

func TestCreateSalt(t *testing.T) {
	salt1, err := CreateSalt()
	if err != nil {
		t.Errorf("Create salt failed. Error: %s", err.Error())
	}
	salt2, err := CreateSalt()
	if err != nil {
		t.Errorf("Create salt failed. Error: %s", err.Error())
	}

	salt1String := hex.EncodeToString(salt1)
	salt2String := hex.EncodeToString(salt2)
	if salt1String == salt2String {
		t.Errorf("Expected 2 diffrent salts, but they are identical. Salt1: %s, Salt2: %s ", salt1String, salt2String)
	}

	salts := [][]byte{salt1, salt2}
	for _, salt := range salts {
		if len(salt) != saltLength {
			t.Errorf("Expected salt length of %d, but got %d", saltLength, len(salt))
		}
	}
}

func TestGenerateApiKey(t *testing.T) {
	key1, err := GenerateApiKey()
	if err != nil {
		t.Errorf("Generate api key failed. Error: %s", err.Error())
	}
	key2, err := GenerateApiKey()
	if err != nil {
		t.Errorf("Generate api key failed. Error: %s", err.Error())
	}

	if key1 == key2 {
		t.Errorf("Expected 2 diffrent keys, but they are identical. Key1: %s, Key2: %s ", key1, key2)
	}

	keys := []string{key1, key2}
	for _, key := range keys {
		if len(key) != apiKeyLength {
			t.Errorf("Expected salt length of %d, but got %d", apiKeyLength, len(key))
		}
	}
}

// TODO: add some more tests
func TestScryptPassword(t *testing.T) {
	salt, err := CreateSalt()
	if err != nil {
		t.Errorf("Create salt failed. Error: %s", err.Error())
	}

	password1, err := ScryptPassword("my-secret-password", salt)
	if err != nil {
		t.Errorf("Sycrpt failed. Error: %s", err.Error())
	}

	password2, err := ScryptPassword("my-secret-password2", salt)
	if err != nil {
		t.Errorf("Sycrpt failed. Error: %s", err.Error())
	}

	if password1 == password2 {
		t.Errorf("Expected two diffrent results, but they are identical. Password1: %s, Password2: %s", password1, password2)
	}
}
