package main

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func validateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return config.Auth.JWTKey, nil
	})
	if err != nil {
		return "", fmt.Errorf("cannot parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		s, assertOK := claims["role"].(string)
		if assertOK {
			return s, nil
		}
	}
	return "", errors.New("cannot get claims map")
}
