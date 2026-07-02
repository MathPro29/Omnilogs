package handler

import (
	"context"

	workerusecase "omnilogs-api/internal/worker/usecase"
)

type Handler struct {
	usecase workerusecase.Usecase
}

func NewHandler(usecase workerusecase.Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) Run(ctx context.Context) error {
	return h.usecase.Run(ctx)
}

func (h *Handler) RunOnce(ctx context.Context) (bool, error) {
	return h.usecase.RunOnce(ctx)
}
