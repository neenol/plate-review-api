package clerk

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Client verifies Clerk-issued JWTs offline using the instance's RSA public key.
type Client struct {
	publicKey *rsa.PublicKey
}

// New parses the PEM-encoded RSA public key and returns a ready Client.
func New(pemKey string) (*Client, error) {
	pemKey = strings.TrimSpace(pemKey)
	// Clerk exports the key with literal \n; normalise both formats.
	pemKey = strings.ReplaceAll(pemKey, `\n`, "\n")

	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, errors.New("clerk: failed to decode PEM block from CLERK_JWT_KEY")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("clerk: parse public key: %w", err)
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("clerk: CLERK_JWT_KEY is not an RSA public key")
	}

	return &Client{publicKey: rsaKey}, nil
}

// VerifyToken parses and validates the JWT, returning the Clerk user ID (sub claim).
func (c *Client) VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("clerk: unexpected signing method %v", t.Header["alg"])
		}
		return c.publicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil {
		return "", fmt.Errorf("clerk: invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("clerk: token claims invalid")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("clerk: missing sub claim")
	}

	return sub, nil
}
