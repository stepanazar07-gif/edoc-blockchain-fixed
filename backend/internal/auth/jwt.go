package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

// Секретный ключ (в реальном проекте вынесите в .env)
var jwtSecret = []byte("your-secret-key-change-in-production")

type Claims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}

// GenerateToken создаёт JWT токен для пользователя
func GenerateToken(userID string) (string, error) {
    claims := Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// ValidateToken проверяет и декодирует JWT токен
func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, jwt.ErrSignatureInvalid
}