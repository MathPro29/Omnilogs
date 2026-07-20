package handler

import (
	"omnilogs-api/configs"
	"omnilogs-api/internal/main_logs/usecase"
)

type Handler struct {
	usecase   usecase.Usecase
	natsQueue *configs.NATSQueue
}

func NewHandler(usecase usecase.Usecase, natsQueue *configs.NATSQueue) *Handler {
	return &Handler{usecase: usecase, natsQueue: natsQueue}
}
