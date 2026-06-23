package service

import "omnilogs-api/utils"

type PasswordService interface {
	Hash(password string) (string, error)
	Check(password string, passwordHash string) bool
}

type BCryptPasswordService struct{}

func NewBCryptPasswordService() *BCryptPasswordService {
	return &BCryptPasswordService{}
}

func (s *BCryptPasswordService) Hash(password string) (string, error) {
	return utils.HashPassword(password)
}

func (s *BCryptPasswordService) Check(password string, passwordHash string) bool {
	return utils.CheckPasswordHash(password, passwordHash)
}
