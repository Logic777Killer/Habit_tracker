package auth

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecretMu sync.RWMutex
	jwtSecret   []byte
)

// SetJWTSecret настраивает секрет подписи JWT из конфигурации приложения.
func SetJWTSecret(secret string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return errors.New("JWT secret is required")
	}

	jwtSecretMu.Lock()
	defer jwtSecretMu.Unlock()
	jwtSecret = []byte(secret)
	return nil
}

func getJWTSecret() ([]byte, error) {
	jwtSecretMu.RLock()
	defer jwtSecretMu.RUnlock()
	if len(jwtSecret) == 0 {
		return nil, errors.New("JWT secret is not configured")
	}

	secret := make([]byte, len(jwtSecret))
	copy(secret, jwtSecret)
	return secret, nil
}

// HashPassword хэширует пароль
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash сравнивает пароль с хэшем
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Claims структура для данных внутри JWT
type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken создает JWT токен
func GenerateToken(userID int, role string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ValidateToken проверяет токен и возвращает UserID
func ValidateToken(tokenString string) (int, string, error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getJWTSecret()
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, "", errors.New("token expired")
		}
		return 0, "", errors.New("invalid token")
	}

	if !tkn.Valid {
		return 0, "", errors.New("invalid token")
	}

	return claims.UserID, claims.Role, nil
}
