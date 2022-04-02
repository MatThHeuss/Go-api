package auth

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var mySignInKey = []byte("value_secret_test")

func CreateToken(user_id string, user_type string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["user_id"] = user_id
	claims["user_type"] = user_type
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()
	tokenString, err := token.SignedString(mySignInKey)

	if err != nil {
		fmt.Errorf("Algo deu errado: %s", err.Error())
		return "", err
	}

	return tokenString, nil

}
