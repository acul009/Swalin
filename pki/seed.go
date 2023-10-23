package pki

import (
	"crypto/rand"
	"errors"
	"os"
	"rahnit-rmm/config"
)

var seed []byte

const seedFilePath = "seed.pem"

func GetSeed() []byte {
	return seed
}

func init() {
	_, err := os.Stat(config.GetFilePath(seedFilePath))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}
		seed := make([]byte, 32)
		if _, err := rand.Read(seed); err != nil {
			panic(err)
		}
		err = SavePasswordToFile(config.GetFilePath(seedFilePath), seed)
		if err != nil {
			panic(err)
		}
	} else {
		seed, err = LoadPasswordFromFile(config.GetFilePath(seedFilePath))
		if err != nil {
			panic(err)
		}
	}
}
