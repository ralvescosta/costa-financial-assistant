package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"

	"crypto/rsa"
)

// extractKID reads the JWT header without full validation to retrieve the key ID.
func extractKID(tokenStr string) (string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("token does not have 3 parts")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("decode header: %w", err)
	}
	var header struct {
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", fmt.Errorf("unmarshal header: %w", err)
	}
	if header.Kid == "" {
		return "", fmt.Errorf("token header missing kid claim")
	}
	return header.Kid, nil
}

// parseAndValidateToken verifies the JWT signature and expiry using the provided RSA public key.
func parseAndValidateToken(tokenStr string, pub *rsa.PublicKey) (*identityv1.JwtClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pub, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type")
	}

	sub, _ := mc["sub"].(string)
	projectID, _ := mc["project_id"].(string)
	role, _ := mc["role"].(string)

	var iat, exp int64
	if v, ok := mc["iat"].(float64); ok {
		iat = int64(v)
	}
	if v, ok := mc["exp"].(float64); ok {
		exp = int64(v)
	}

	return &identityv1.JwtClaims{
		Subject:   sub,
		ProjectId: projectID,
		Role:      role,
		IssuedAt:  iat,
		ExpiresAt: exp,
	}, nil
}
