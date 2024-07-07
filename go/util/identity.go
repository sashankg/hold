package util

import (
	"bufio"
	"io"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
)

func LoadIdentity(path string) (crypto.PrivKey, error) {
	file, err := os.Open(path)
	if err != nil {
		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		privKey, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
		privKeyBytes, err := crypto.MarshalPrivateKey(privKey)
		if err != nil {
			return nil, err
		}
		_, err = file.Write(privKeyBytes)
		if err != nil {
			return nil, err
		}
		return privKey, nil
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	privKeyBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPrivateKey(privKeyBytes)
}
