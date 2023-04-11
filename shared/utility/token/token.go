package token

import (
	"errors"
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/robinmin/xally/shared/model"
	log "github.com/sirupsen/logrus"
)

func GenerateAccessToken(app_token string, email string) (string, error) {
	user_info, err := model.NewUserInfo(app_token, email, "")
	if err != nil {
		log.Error("Failed to get current user information: %v", err.Error())
		return "", err
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = true

	claims["username"] = user_info.Username
	claims["uid"] = user_info.UserID
	claims["hostname"] = user_info.Hostname
	claims["email"] = email
	claims["device_info"] = user_info.DeviceInfo

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(app_token))
}

func ExtractAccessInfo(app_key string, access_token string) (jwt.MapClaims, error) {
	if access_token == "" {
		log.Error("Blank access token in ExtractAccessInfo")
		return nil, nil
	}

	token, err := jwt.Parse(access_token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(app_key), nil
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("Invalid access token")
}
