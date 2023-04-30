package utils

import (
	_const "chat-ai/chat-server/const"
	mylog "chat-ai/chat-server/log"
	"chat-ai/chat-server/model"
	"chat-ai/chat-server/repo"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"time"
)

func GenerateRefreshToken(userID uint, email string) (string, error) {
	// create claims
	claims := jwt.MapClaims{
		"userID": userID,
		"email":  email,
		"exp":    time.Now().Add(time.Hour * 24 * 30 * 6).Unix(), // token expiration time (30 days)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// generate encoded token using secret key
	secretKey := []byte(_const.REFRESH_TOKEN_SECRET_KEY)
	return token.SignedString(secretKey)
}

func RefreshToken(c *gin.Context, accessToken, refreshTokenStr string) (string, error) {
	if refreshTokenStr == "" {
		return "", fmt.Errorf("refresh token is required")
	}
	// Verify and extract claims from the refresh token
	refreshClaims := jwt.MapClaims{}
	refreshToken, err := jwt.ParseWithClaims(refreshTokenStr, refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(_const.REFRESH_TOKEN_SECRET_KEY), nil
	})
	if err != nil {
		return "", fmt.Errorf("invalid refresh token")
	}
	if !refreshToken.Valid {
		return "", fmt.Errorf("refresh token is not valid")
	}
	userID := int(refreshClaims["userID"].(float64))
	email := refreshClaims["email"].(string)

	// Check if user exists and retrieve the latest user information
	var user = &model.User{}
	result := repo.QueryById(userID, user)
	if result.Error != nil {
		return "", fmt.Errorf("failed to retrieve user information")
	}
	if user == nil || user.Email != email || user.Status == _const.USER_BANNED_STATUS {
		mylog.Logger.Infof("用户异常: %+v", user)
		return "", fmt.Errorf("user not authorized")
	}

	// Generate a new access token using the user information
	accessToken, err = GenerateJWT(user.ID, user.Email)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token")
	}
	c.Set("userID", uint(userID))
	mylog.Logger.Infof("当前操作的用户id: %d", uint(userID))
	return accessToken, nil
}

func GenerateJWT(userID uint, email string) (string, error) {
	// create claims
	claims := jwt.MapClaims{
		"userID": userID,
		"email":  email,
		"exp":    time.Now().Add(time.Hour * 12).Unix(), // token expiration time
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// generate encoded token using secret key
	secretKey := []byte(_const.ACCESS_TOKEN_SECRET_KEY) // replace with your own secret key
	return token.SignedString(secretKey)
}

func ParseJWT(tokenString string) (jwt.MapClaims, error) {
	// decode token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		// secret key
		secretKey := []byte(_const.ACCESS_TOKEN_SECRET_KEY) // replace with your own secret key
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// validate claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}
