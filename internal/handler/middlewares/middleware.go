package middlewares

import (
	"net/http"

	"github.com/warehouse/ai-service/internal/adapter/auth"
	"github.com/warehouse/ai-service/internal/config"
	"github.com/warehouse/ai-service/internal/domain"
	"github.com/warehouse/ai-service/internal/pkg/logger"
)

const (
	VersionDelimiter = ":" // Разделитель составных частей версий
	VersionHeader    = "Coffee-Version"

	AuthHeader    = "Authorization"
	TokenStart    = "Bearer "       // Префикс значения заголовка с авторизацией
	TokenStartInd = len(TokenStart) // Индекс, с которого в заголовке авторизации должен начинаться jwt токен

	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
)

type (
	Middleware interface {
		JwtAuthMiddleware(purpose domain.AuthPurpose) func(http.Handler) http.Handler
		QueueMiddleware(h http.Handler) http.Handler
	}

	middleware struct {
		log logger.Logger

		timeouts    config.Timeouts
		authAdapter auth.Adapter
		queue       chan struct{}
	}
)

func NewMiddleware(
	log logger.Logger,
	timeouts config.Timeouts,
	authAdapter auth.Adapter,
) Middleware {
	return &middleware{
		log:         log,
		timeouts:    timeouts,
		authAdapter: authAdapter,
	}
}
