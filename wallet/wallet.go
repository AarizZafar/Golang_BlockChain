package wallet

import "crypto/ecdsa"

type Wallet struct {
	privateKey *ecdsa.PrivateKey
	publicKey *ecdsa.PublicKey
}

func NewWallet() *Wallet {
	w := new(Wallet)
}