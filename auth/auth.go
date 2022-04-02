package auth

import (
	"fmt"
	"net/http"
	"strings"
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
		err := fmt.Errorf("algo deu errado: %s", err.Error())
		return "", err
	}

	return tokenString, nil

}

func VerifyToken(r *http.Request) (jwt.Claims, error) {
	bearerToken := r.Header.Get("Authorization")
	jwtToken := strings.Split(bearerToken, " ")[1]

	token, err := jwt.Parse(jwtToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return mySignInKey, nil
	})

	fmt.Println(token)

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, nil
}
