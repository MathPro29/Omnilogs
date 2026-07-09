package handler

import (
	authusecase "omnilogs-api/internal/auth/usecase"
)

type Handler struct {
	usecase authusecase.Usecase
}

func NewHandler(usecase authusecase.Usecase) *Handler {
	return &Handler{usecase: usecase}
}