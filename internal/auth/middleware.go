package auth

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spendalt/backend/internal/lang"
)

func AuthRequired(secret string, tokenStore TokenStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).JSON(fiber.Map{"error": lang.ErrUnauthorized})
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": lang.ErrInvalidToken})
		}

		claims := token.Claims.(jwt.MapClaims)

		if jti, ok := claims["jti"].(string); ok && jti != "" {
			revoked, err := tokenStore.IsRevoked(jti)
			if err != nil || revoked {
				return c.Status(401).JSON(fiber.Map{"error": lang.ErrTokenRevoked})
			}
		}

		c.Locals("user_id", claims["user_id"].(string))
		c.Locals("token", tokenString)
		return c.Next()
	}
}
