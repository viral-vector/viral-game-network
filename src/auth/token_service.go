package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strconv"
	"strings"
	"time"
	"viral-game-network/src/database/repository"
)

type TokenAudience string

const (
	UserAudience  TokenAudience = "vgn:user"
	AdminAudience TokenAudience = "vgn:admin"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// VNET_KEY is shared with game clients and pods. It must never sign sessions.
func signingKey() ([]byte, error) {
	key := os.Getenv("VNET_TOKEN_KEY")
	if len(key) < 32 {
		return nil, fmt.Errorf("VNET_TOKEN_KEY must contain at least 32 bytes")
	}
	if key == os.Getenv("VNET_KEY") {
		return nil, fmt.Errorf("VNET_TOKEN_KEY must differ from VNET_KEY")
	}
	return []byte(key), nil
}

func ValidateSigningKey() error {
	_, err := signingKey()
	return err
}

func validIdentity(subject string, audience TokenAudience) bool {
	table := ""
	switch audience {
	case UserAudience:
		table = "User:"
	case AdminAudience:
		table = "Admin:"
	default:
		return false
	}
	return strings.HasPrefix(subject, table) && len(subject) > len(table)
}

func GenerateToken(username, subject string, audience TokenAudience) (string, error) {
	key, err := signingKey()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(username) == "" || !validIdentity(subject, audience) {
		return "", fmt.Errorf("invalid session identity")
	}
	configs, err := repository.GetSystemConfigs()
	if err != nil {
		return "", fmt.Errorf("load session configuration: %w", err)
	}
	appName, expiry := os.Getenv("VNET_NAME"), os.Getenv("VNET_TOKEN_EXPIRE")
	for _, config := range configs {
		if config.Val == "" {
			continue
		}
		switch config.Key {
		case "VNET_NAME":
			appName = config.Val
		case "VNET_TOKEN_EXPIRE":
			expiry = config.Val
		}
	}
	if appName == "" {
		appName = "VNet"
	}
	minutes, err := strconv.ParseInt(expiry, 10, 64)
	if err != nil || minutes <= 0 || minutes > int64((1<<63-1)/time.Minute) {
		return "", fmt.Errorf("invalid token lifetime in minutes")
	}
	now := time.Now()
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Audience:  jwt.ClaimStrings{string(audience)},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(minutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    appName,
			ID:        uuid.NewString(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}

func ValidateToken(tokenString string) (*Claims, error) {
	key, err := signingKey()
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || strings.TrimSpace(claims.Username) == "" || len(claims.Audience) != 1 || !validIdentity(claims.Subject, TokenAudience(claims.Audience[0])) {
		return nil, fmt.Errorf("invalid session claims")
	}
	return claims, nil
}

func ValidateTokenFor(tokenString string, audience TokenAudience) (*Claims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Audience[0] != string(audience) {
		return nil, fmt.Errorf("invalid session audience")
	}
	return claims, nil
}

func GenerateAppKey(size int32) string {
	key := strings.Replace(uuid.New().String(), "-", "", -1)
	return key[0:size]
}

func ValidateAppKey(key string) (bool, error) {
	keys, err := repository.GetApiKeys()
	if err != nil {
		return false, fmt.Errorf("Error pulling keys")
	}
	hashedExtKey := sha256.Sum256([]byte(key))
	for _, k := range keys {
		hashedAppKey := sha256.Sum256([]byte(k.Val))
		if subtle.ConstantTimeCompare(hashedAppKey[:], hashedExtKey[:]) == 1 {
			return true, nil
		}
	}
	return false, fmt.Errorf("Invalid app key")
}

func HashGenerate(item string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(item), 14)
	return string(bytes), err
}

func HashValidate(item string, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(item))
	return err == nil, err
}
