package service

import (
	"crypto/rand"
	"encoding/hex"
)

type OpaqueTokenService interface {
	NewToken(byteLength int) (string, error)
}

type RandomOpaqueTokenService struct{}

func NewRandomOpaqueTokenService() *RandomOpaqueTokenService {
	return &RandomOpaqueTokenService{}
}

func (s *RandomOpaqueTokenService) NewToken(byteLength int) (string, error) {
	value := make([]byte, byteLength)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
