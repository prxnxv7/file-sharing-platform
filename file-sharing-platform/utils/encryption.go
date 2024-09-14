package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"os"
)

// EncryptFile encrypts the file before saving it locally
func EncryptFile(filepath, key string) error {
	// Open the file to be encrypted
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new AES cipher using the key
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return err
	}

	// Create a new file to save the encrypted data
	outFile, err := os.Create(filepath + ".enc")
	if err != nil {
		return err
	}
	defer outFile.Close()

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	writer := &cipher.StreamWriter{S: stream, W: outFile}

	if _, err = io.Copy(writer, file); err != nil {
		return err
	}

	return nil
}

// DecryptFile decrypts the file for reading
func DecryptFile(filepath, key string) error {
	// Open the encrypted file
	file, err := os.Open(filepath + ".enc")
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new AES cipher using the key
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return err
	}

	// Create a file to save the decrypted data
	outFile, err := os.Create(filepath + ".dec")
	if err != nil {
		return err
	}
	defer outFile.Close()

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewCFBDecrypter(block, iv)
	reader := &cipher.StreamReader{S: stream, R: file}

	if _, err = io.Copy(outFile, reader); err != nil {
		return err
	}

	return nil
}
