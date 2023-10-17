package middlewares

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/prizmsol/prizmsol-server/database"
)

type authString string

type ErrorPayload struct {
	Message string
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		if auth == "" {
			next.ServeHTTP(w, r)
			return
		}

		bearer := "Bearer "
		auth = auth[len(bearer):]

		validate, err := database.JwtValidate(context.Background(), auth)
		if err != nil || !validate.Valid {
			p := ErrorPayload{
				Message: "Invalid token",
			}
			json.NewEncoder(w).Encode(p)
			return
		}

		customClaim, _ := validate.Claims.(*database.CustomClaim)
		ctx := context.WithValue(r.Context(), authString("auth"), customClaim)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func CtxValue(ctx context.Context) *database.CustomClaim {
	raw, _ := ctx.Value(authString("auth")).(*database.CustomClaim)
	return raw
}
