package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type auther struct {
	secret []byte
	issuer string
}

type claims struct {
	jwt.RegisteredClaims
}

func NewAuther(secret, issuer string) (*auther, error) {
	if secret == "" {
		return nil, fmt.Errorf("missing jwt secret")
	}

	if issuer == "" {
		return nil, fmt.Errorf("missing jwt issuer")
	}

	auther := &auther{
		secret: []byte(secret),
		issuer: issuer,
	}

	return auther, nil
}

// parseJwtToken checks validity of token and returns jwt subject.
// Validty is checked for HS256 algorithm.
func (a *auther) parseJwtToken(token string) (string, error) {
	t, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return a.secret, nil
	})

	if err != nil {
		return "", fmt.Errorf("error validating token: %v", err)
	}

	if claims, ok := t.Claims.(*claims); ok {
		return claims.Subject, nil
	}

	return "", fmt.Errorf("error parsing token: %v", err)
}

// withAuth is a middleware that ensures the request is authenticated before allowing access to the next handler.
// It checks the presence and validity of the Authorization header, expecting a Bearer token format.
// If the Authorization header is missing, invalid, or the JWT token is not valid, it responds with a 401 Unauthorized error.
// If the token is valid, it proceeds to the next handler.
func withAuth(c *APIController, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		const prefix = "Bearer "

		if !strings.HasPrefix(authHeader, prefix) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, prefix)
		if token == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		_, err := c.auther.parseJwtToken(token)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next(w, r)
	}

	return fn
}
