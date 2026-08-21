package handler

import (
	"omnilogs-api/internal/main_logs/usecase"
)

type Handler struct {
	usecase usecase.Usecase
}

func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{usecase: usecase}
}
