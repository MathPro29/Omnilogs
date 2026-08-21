package handler

import (
	"omnilogs-api/internal/product/usecase"
)

type Handler struct{ usecase usecase.Usecase }

func NewHandler(usecase usecase.Usecase) *Handler { return &Handler{usecase: usecase} }
