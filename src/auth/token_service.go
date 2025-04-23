package auth

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"strings"
	"crypto/sha256"
    "crypto/subtle"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"viral-game-network/src/database/repository"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var vnet_key = []byte(os.Getenv("VNET_KEY"))

func GenerateToken(username string) (string, error) {
	app_name := *repository.GetConfigValue("VNET_NAME")
	exp_time := *repository.GetConfigValue("VNET_TOKEN_EXPIRE")

	if app_name == "" {
		app_name = "VNet"
	}
	if exp_time == "" {
		exp_time = os.Getenv("VNET_TOKEN_EXPIRE")
	}

	i, err := strconv.ParseInt(exp_time, 10, 64)
	if err != nil {
  		return "", fmt.Errorf("Parsing Expire Time: %s", err)
	}
	l := time.Duration(+int(i))

	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(l * time.Minute)),
			Issuer: app_name,
		},
	}
	access_token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

 	return access_token.SignedString(vnet_key)
}

func ValidateToken(tokenString string) (*Claims, error) {
	access_token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return vnet_key, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := access_token.Claims.(*Claims)

	if !access_token.Valid || !ok{
		return nil, fmt.Errorf("Invalid Token")
	}
	  
	return claims, err
}

func GenerateAppKey(size int32) string {
	key := strings.Replace(uuid.New().String(), "-", "", -1)
	return key[0:size]
}

func ValidateAppKey(key string) (bool, error) {
	hashedAppKey := sha256.Sum256(vnet_key)
    hashedExtKey := sha256.Sum256([]byte(key))

    if subtle.ConstantTimeCompare(hashedAppKey[:], hashedExtKey[:]) == 1 {
        return true, nil
    }
    return false, fmt.Errorf("Invalid App Key")
}

func HashGenerate(item string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(item), 14)
    return string(bytes), err
}

func HashValidate(item string, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(item))
    return err == nil, err
}