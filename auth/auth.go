package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var mySignInKey = []byte("value_secret_test")

type customClaims struct {
	UserId    string `json:"user_id"`
	User_type string `json:"user_type"`
	Email     string `json:"email"`
	jwt.StandardClaims
}

func CreateToken(user_id string, user_type string, user_email string) (string, error) {

	claims := customClaims{
		UserId:    user_id,
		User_type: user_type,
		Email:     user_email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
			Issuer:    "matheusWebsite",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(mySignInKey)

	if err != nil {
		err := fmt.Errorf("algo deu errado: %s", err.Error())
		return "", err
	}

	return tokenString, nil

}

func VerifyToken(r *http.Request) (string, string, string, int64, error) {
	bearerToken := r.Header.Get("Authorization")
	jwtToken := strings.Split(bearerToken, " ")[1]

	token, err := jwt.ParseWithClaims(jwtToken, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySignInKey, nil
	})

	if err != nil {
		return "", "", "", 0, err
	}

	if claims, ok := token.Claims.(*customClaims); ok && token.Valid {
		return claims.UserId, claims.User_type, claims.Email, claims.ExpiresAt, nil
	}

	return "", "", "", 0, nil
}
