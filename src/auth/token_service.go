package auth

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"crypto/sha256"
    "crypto/subtle"
	"github.com/golang-jwt/jwt/v5"
	"viral-game-network/src/database/type"
)

type Claims struct {
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

var app_key = []byte(os.Getenv("APP_KEY"))

func GenerateToken(user *dbtype.User) (string, error) {
	i, err := strconv.ParseInt(os.Getenv("APP_TOKEN_EXPIRE"), 10, 64)
	if err != nil {
  		return "", fmt.Errorf("Server Error Parsing Expire Time")
	}
	l := time.Duration(+int(i))

	claims := Claims{
		UserName: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(l * time.Minute)),
			Issuer: os.Getenv("APP_NAME"),
		},
	}
	access_token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

 	return access_token.SignedString(app_key)
}

func ValidateToken(tokenString string) (*dbtype.User, error) {
	access_token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return app_key, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := access_token.Claims.(*Claims)

	if !access_token.Valid || !ok{
		return nil, fmt.Errorf("Invalid Token")
	}

	user := new(dbtype.User)
	user.Name = claims.UserName
	  
	return user, err
}

func ValidateAppKey(key string) (bool, error) {
	hashedAppKey := sha256.Sum256(app_key)
    hashedExtKey := sha256.Sum256([]byte(key))

    if subtle.ConstantTimeCompare(hashedAppKey[:], hashedExtKey[:]) == 1 {
        return true, nil
    }
    return false, fmt.Errorf("Invalid App Key")
}