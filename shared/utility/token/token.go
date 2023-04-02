package token

import (
	"errors"
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/robinmin/xally/shared/model"
	log "github.com/sirupsen/logrus"
)

// type Token struct {
// 	ApiSecret     string // your own secret string for signing the token
// 	TokenLifespan uint32 // how long each token will last (hour)
// }

// func (t *Token) GenerateToken(user_id uint) (string, error) {
// 	claims := jwt.MapClaims{}
// 	claims["authorized"] = true
// 	claims["user_id"] = user_id
// 	claims["exp"] = time.Now().Add(time.Hour * time.Duration(t.TokenLifespan)).Unix()
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	return token.SignedString([]byte(t.ApiSecret))
// }

// func (t *Token) GetUserIDFromToken(token_str string) (uint, error) {
// 	// 解析token
// 	token, err := jwt.Parse(token_str, func(token *jwt.Token) (interface{}, error) {
// 		// 根据实际情况配置JWT密钥
// 		return []byte(t.ApiSecret), nil
// 	})
// 	if err != nil || !token.Valid {
// 		return 0, err
// 	}

// 	// 将用户id存入上下文中
// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok {
// 		return 0, errors.New("无效的token : 类型不对")
// 	}

// 	user_id, ok := claims["user_id"].(uint)
// 	if !ok {
// 		return 0, errors.New("无效的token : user_id")
// 	}
// 	return user_id, nil
// }

// func (t *Token) TokenValid(c *gin.Context) error {
// 	tokenString := t.ExtractToken(c)
// 	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
// 		}
// 		return []byte(t.ApiSecret), nil
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (t *Token) ExtractToken(c *gin.Context) string {
// 	token := c.Query("token")
// 	if token != "" {
// 		return token
// 	}
// 	bearerToken := c.Request.Header.Get("Authorization")
// 	if len(strings.Split(bearerToken, " ")) == 2 {
// 		return strings.Split(bearerToken, " ")[1]
// 	}
// 	return ""
// }

// func (t *Token) ExtractTokenID(c *gin.Context) (uint, error) {
// 	tokenString := t.ExtractToken(c)
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
// 		}
// 		return []byte(t.ApiSecret), nil
// 	})
// 	if err != nil {
// 		return 0, err
// 	}
// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if ok && token.Valid {
// 		uid, err := strconv.ParseUint(fmt.Sprintf("%.0f", claims["user_id"]), 10, 32)
// 		if err != nil {
// 			return 0, err
// 		}
// 		return uint(uid), nil
// 	}
// 	return 0, nil
// }

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
