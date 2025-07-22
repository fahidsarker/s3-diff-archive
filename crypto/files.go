package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"os"
)

// Example function to derive a key from a string
func deriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

func EncryptFile(sKey string, inputPath, outputPath string) error {
	key := deriveKey(sKey)
	// Open input file
	plainFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer plainFile.Close()

	// Create output file
	encFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer encFile.Close()

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Use AES-GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Create nonce (12 bytes for GCM)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// Write nonce to the beginning of the output file
	if _, err := encFile.Write(nonce); err != nil {
		return err
	}

	// Read all data to encrypt
	plainData, err := io.ReadAll(plainFile)
	if err != nil {
		return err
	}

	// Encrypt
	cipherData := aesGCM.Seal(nil, nonce, plainData, nil)

	// Write encrypted data
	if _, err := encFile.Write(cipherData); err != nil {
		return err
	}

	return nil
}

func DecryptFile(sKey string, inputPath, outputPath string) error {
	key := deriveKey(sKey)
	// Open encrypted file
	encFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer encFile.Close()

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Use AES-GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := aesGCM.NonceSize()

	// Read nonce (first bytes of file)
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(encFile, nonce); err != nil {
		return err
	}

	// Read rest of file (ciphertext)
	cipherData, err := io.ReadAll(encFile)
	if err != nil {
		return err
	}

	// Decrypt
	plainData, err := aesGCM.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return err
	}

	// Write decrypted data to output file
	plainFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer plainFile.Close()

	if _, err := plainFile.Write(plainData); err != nil {
		return err
	}

	return nil
}
