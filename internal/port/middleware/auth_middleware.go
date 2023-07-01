// Package middleware http server middlewares
package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	// User middleware key-value for user id
	User contextKey = "user"
)

// CustomClaims custom JWT claims
type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Authenticator JWT authenication middleware
type Authenticator struct {
	publicKey *ecdsa.PublicKey
}

// NewAuthenticator create new authenticator instance
func NewAuthenticator() (*Authenticator, error) {

	p := os.Getenv("JWT_PRIVATE_KEY")
	if p == "" {
		return nil, errors.New("$JWT_PRIVATE_KEY is not set")
	}

	bb, e := os.ReadFile(filepath.Clean(p))
	if e != nil {
		return nil, fmt.Errorf("unable to open private key %v", e)
	}

	b, _ := pem.Decode(bb)

	pk, e := x509.ParseECPrivateKey(b.Bytes)
	if e != nil {
		return nil, fmt.Errorf("failed to parse ecdsa private key %v", e)
	}

	return &Authenticator{
		publicKey: &pk.PublicKey,
	}, nil
}

// ValidateJWT parse and validate JWT token
func (a *Authenticator) ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// get authorization header value
		h := r.Header.Get("Authorization")

		t := strings.Split(h, "Bearer: ")[1]

		tk, e := jwt.ParseWithClaims(t, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method %s", token.Header["alg"])
			}

			return a.publicKey, nil
		})
		if e != nil {
			http.Error(w, "invalid auth token", http.StatusUnauthorized)
			return
		}

		if c, ok := tk.Claims.(*CustomClaims); ok && tk.Valid {
			ctx := context.WithValue(r.Context(), User, c.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
		}
	})
}
