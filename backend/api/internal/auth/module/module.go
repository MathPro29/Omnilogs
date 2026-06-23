package auth

import (
	"time"

	"omnilogs-api/configs"
	"omnilogs-api/internal/auth/handler"
	"omnilogs-api/internal/auth/repository"
	"omnilogs-api/internal/auth/service"
	"omnilogs-api/internal/auth/usecase"

	"gorm.io/gorm"
)

func NewHandler(db *gorm.DB, env *configs.Env) *handler.Handler {
	repo := repository.NewRepository(db)
	tokenManager := service.NewJWTTokenManager(env.JWTSecret)
	passwordService := service.NewBCryptPasswordService()
	opaqueTokenService := service.NewRandomOpaqueTokenService()

	authUsecase := usecase.NewUsecase(
		repo,
		tokenManager,
		passwordService,
		opaqueTokenService,
		time.Duration(env.AccessTokenExpireSeconds)*time.Second,
		time.Duration(env.RefreshTokenExpireSeconds)*time.Second,
	)

	return handler.NewHandler(authUsecase)
}
