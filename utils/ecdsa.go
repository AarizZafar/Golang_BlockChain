package utils

import (
	"fmt"
	"math/big"
)

// r, s value are part of the digital signature
type Signature struct {
	R *big.Int
	S *big.Int
}

func (s *Signature) string() string {
	return fmt.Sprintf("%x%x", s.R, s.S)
}