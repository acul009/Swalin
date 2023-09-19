package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"golang.org/x/crypto/argon2"
)

// EncryptDataWithPassword encrypts data using a password and returns the encrypted result.
func EncryptDataWithPassword(password []byte, data []byte) ([]byte, error) {
	// Generate a random salt
	salt := make([]byte, aes.BlockSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Derive an encryption key from the password and salt
	key, err := deriveKeyFromPassword(password, salt)
	if err != nil {
		return nil, err
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a stream cipher for encryption
	stream := cipher.NewCFBEncrypter(block, salt)

	// Encrypt the data
	encryptedData := make([]byte, len(data))
	stream.XORKeyStream(encryptedData, data)

	// Prepend the salt to the encrypted data
	encryptedDataWithSalt := append(salt, encryptedData...)

	return encryptedDataWithSalt, nil
}

// DecryptDataWithPassword decrypts data that was encrypted with a password.
func DecryptDataWithPassword(password, encryptedData []byte) ([]byte, error) {
	// Extract the salt from the beginning of the encrypted data
	salt := encryptedData[:aes.BlockSize]
	encryptedPayload := encryptedData[aes.BlockSize:]

	// Derive the encryption key from the password and salt
	key, err := deriveKeyFromPassword(password, salt)
	if err != nil {
		return nil, err
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a stream cipher for decryption
	stream := cipher.NewCFBDecrypter(block, salt)

	// Decrypt the payload
	decryptedData := make([]byte, len(encryptedPayload))
	stream.XORKeyStream(decryptedData, encryptedPayload)

	return decryptedData, nil
}

func deriveKeyFromPassword(password []byte, salt []byte) ([]byte, error) {
	// Parameters for Argon2id (adjust according to your security requirements)
	timeCost := 1           // Number of iterations
	memoryCost := 64 * 1024 // Memory usage in KiB
	parallelism := 4        // Number of threads
	keyLen := 32            // Desired key length in bytes

	key := argon2.IDKey(password, salt, uint32(timeCost), uint32(memoryCost), uint8(parallelism), uint32(keyLen))
	return key, nil
}
