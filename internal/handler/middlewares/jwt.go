package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/warehouse/ai-service/internal/domain"
	"github.com/warehouse/ai-service/internal/handler/converters"
	"github.com/warehouse/ai-service/internal/handler/writers"
	"github.com/warehouse/ai-service/internal/warehousepb"
)

func (m *middleware) JwtAuthMiddleware(purpose domain.AuthPurpose) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a := r.Header.Get(AuthHeader)
			isToken := strings.HasPrefix(a, TokenStart)
			if !isToken {
				next.ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), m.timeouts.AuthTimeout)
			defer cancel()

			acc, num, err := m.authAdapter.Authenticate(
				ctx,
				&warehousepb.AuthRequest{Token: a[TokenStartInd:], Purpose: int64(purpose)},
			)
			if err != nil {
				details := ""
				if err.Details != nil {
					details = err.Details.Error()
				}
				m.log.Zap().Error(fmt.Sprintf("auth_failed_log_jwt err=%s, token=%s", fmt.Sprintf("(reason=%s, details=%s)", err.Reason, details), a[TokenStartInd:]))
				writers.SendJSON(w, 200, converters.MakeJsonErrorResponseWithErrorsError(err))
				return
			}
			ctx = context.WithValue(r.Context(), domain.AccountCtxKey, &acc)
			ctx = context.WithValue(ctx, domain.TokenNumberCtxKey, num)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
