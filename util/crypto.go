package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/argon2"

	"math/big"
	pseudorand "math/rand"
)

var encryptionInfoDelimiter = []byte("\n")

// EncryptDataWithPassword encrypts data using a password and returns the encrypted result.
func EncryptDataWithPassword(password []byte, data []byte) ([]byte, error) {
	parameters, err := generateEncryptionParameters()
	if err != nil {
		return nil, fmt.Errorf("failed generating encryption parameters: %w", err)
	}

	// Derive an encryption key from the password and salt
	key, err := deriveKeyFromPassword(password, parameters.ArgonParameters)
	if err != nil {
		return nil, fmt.Errorf("failed deriving encryption key: %w", err)
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed creating cipher: %w", err)
	}

	// Create a stream cipher for encryption
	stream := cipher.NewCFBEncrypter(block, parameters.IV)

	// Encrypt the data
	encryptedData := make([]byte, len(data))
	stream.XORKeyStream(encryptedData, data)

	// Prepend the parameters to the encrypted data
	encryptionInfo, err := json.Marshal(parameters)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling parameters: %w", err)
	}
	encryptionInfo = append(encryptionInfo, encryptionInfoDelimiter...)

	dataToSave := append(encryptionInfo, encryptedData...)

	return dataToSave, nil
}

// DecryptDataWithPassword decrypts data that was encrypted with a password.
func DecryptDataWithPassword(password, encryptedData []byte) ([]byte, error) {

	// Extract the parameters from the beginning of the encrypted data
	parts := bytes.SplitN(encryptedData, encryptionInfoDelimiter, 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed splitting encrypted data")
	}

	encodedParams := parts[0]
	encryptedPayload := parts[1]

	parameters := &EncryptionParameters{}
	err := json.Unmarshal(encodedParams, parameters)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling parameters: %w", err)
	}

	// Derive the encryption key from the password and salt
	key, err := deriveKeyFromPassword(password, parameters.ArgonParameters)
	if err != nil {
		return nil, fmt.Errorf("failed deriving encryption key: %w", err)
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed creating cipher: %w", err)
	}

	// Create a stream cipher for decryption
	stream := cipher.NewCFBDecrypter(block, parameters.IV)

	// Decrypt the payload
	decryptedData := make([]byte, len(encryptedPayload))
	stream.XORKeyStream(decryptedData, encryptedPayload)

	return decryptedData, nil
}

type ArgonOptions struct {
	TimeCost    uint32
	MemoryCost  uint32
	Parallelism uint8
	KeyLength   uint32
}

type ArgonParameters struct {
	ArgonOptions
	Salt []byte
}

type EncryptionParameters struct {
	ArgonParameters
	IV []byte
}

type argonStrength byte

const (
	ArgonStrengthDefault argonStrength = iota
	ArgonStrengthStrong  argonStrength = iota
)

func GenerateArgonParameters(strength argonStrength) (ArgonParameters, error) {
	// Generate a random salt
	salt := make([]byte, aes.BlockSize)
	if _, err := rand.Read(salt); err != nil {
		return ArgonParameters{}, fmt.Errorf("failed generating salt: %w", err)
	}

	return ArgonParameters{
		ArgonOptions: defaultArgonOptions(strength),
		Salt:         salt,
	}, nil
}

func GenerateDecoyArgonParametersFromSeed(seed []byte, salt []byte) (ArgonParameters, error) {
	// Generate a random salt
	key, err := deriveKeyFromPassword(seed, ArgonParameters{
		ArgonOptions: defaultArgonOptions(ArgonStrengthDefault),
		Salt:         salt,
	})

	if err != nil {
		return ArgonParameters{}, fmt.Errorf("failed generating argon parameters: %w", err)
	}

	seedInt := binary.LittleEndian.Uint64(key[:8])

	r := pseudorand.New(pseudorand.NewSource(int64(seedInt)))

	decoySalt := make([]byte, aes.BlockSize)
	if _, err := r.Read(salt); err != nil {
		return ArgonParameters{}, fmt.Errorf("failed generating salt: %w", err)
	}

	return ArgonParameters{
		ArgonOptions: defaultArgonOptions(ArgonStrengthStrong),
		Salt:         decoySalt,
	}, nil

}

func (p ArgonParameters) IsInsecure() bool {
	min := defaultArgonOptions(ArgonStrengthDefault)
	if p.TimeCost < min.TimeCost {
		return true
	}

	if p.MemoryCost < min.MemoryCost {
		return true
	}

	if p.Parallelism < min.Parallelism {
		return true
	}

	if p.KeyLength < min.KeyLength {
		return true
	}

	return false
}

func generateEncryptionParameters() (EncryptionParameters, error) {
	// Generate a random iv
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return EncryptionParameters{}, fmt.Errorf("failed generating iv: %w", err)
	}

	parameters, err := GenerateArgonParameters(ArgonStrengthStrong)
	if err != nil {
		return EncryptionParameters{}, fmt.Errorf("failed generating argon parameters: %w", err)
	}

	return EncryptionParameters{
		ArgonParameters: parameters,
		IV:              iv,
	}, nil
}

func defaultArgonOptions(strength argonStrength) ArgonOptions {
	switch strength {
	case ArgonStrengthStrong:
		return ArgonOptions{
			TimeCost:    4,
			MemoryCost:  1024 * 1024,
			Parallelism: 8,
			KeyLength:   32,
		}
	default:
		return ArgonOptions{
			TimeCost:    1,
			MemoryCost:  64 * 1024,
			Parallelism: 4,
			KeyLength:   32,
		}
	}
}

func HashPassword(password []byte, params ArgonParameters) ([]byte, error) {
	return deriveKeyFromPassword(password, params)
}

func VerifyPassword(password []byte, hash []byte, params ArgonParameters) error {
	actualHash, err := deriveKeyFromPassword(password, params)
	if err != nil {
		return err
	}

	if !bytes.Equal(actualHash, hash) {
		return fmt.Errorf("hashes do not match")
	}

	return nil
}

func deriveKeyFromPassword(password []byte, params ArgonParameters) ([]byte, error) {

	key := argon2.IDKey(password, params.Salt, uint32(params.TimeCost), uint32(params.MemoryCost), uint8(params.Parallelism), uint32(params.KeyLength))
	return key, nil
}

const passwordLength = 64
const passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
const charsetLength = len(passwordCharset)

func GeneratePassword() ([]byte, error) {
	// Create a byte slice to hold the random password
	password := make([]byte, passwordLength)

	max := big.NewInt(int64(charsetLength))

	for i := 0; i < passwordLength; i++ {
		// Generate a random index within the character set length
		index, err := rand.Int(rand.Reader, max)
		if err != nil {
			return []byte{}, err
		}

		// Use the index to select a character from the character set
		password[i] = passwordCharset[index.Int64()]
	}

	return password, nil
}
