package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResult struct {
	Token string `json:"token"`
}

func generateToken(username string) (string, error) {
	role := "admin"
	if config.App.MapDatabase == "" {
		role = "viewer"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":  username,
			"nbf":  time.Now().Add(-5 * time.Minute).Unix(),
			"exp":  time.Now().Add(time.Duration(config.Auth.TTL) * time.Second).Unix(),
			"role": role,
		})

	tokenString, err := token.SignedString([]byte(config.Auth.JWTKey))
	if err != nil {
		return "", fmt.Errorf("cannot generate token: %w", err)
	}
	return tokenString, nil
}

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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var item loginRequest
	req := PrepareReq(w, r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("cannot read body", "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "0000", "Cannot read request", err)
		return
	}
	err = json.Unmarshal(body, &item)
	if err != nil {
		log.Error("could not decode login", "error", err)
		req.ReturnError(w, http.StatusBadRequest, "invalid_request", "JSON parse error", err)
		return
	}
	if item.Username != "admin" {
		log.Error("invalid user", "username", item.Username)
		req.ReturnError(w, http.StatusBadRequest, "invalid_request", "Invalid user, only user admin is supported in this model", nil)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(config.Auth.AdminPassword), []byte(item.Password))
	if err != nil {
		log.Error("cannot compare bcrypt hash", "error", err)
		req.ReturnError(w, http.StatusForbidden, "invalid_password", "Invalid password", nil)
		return
	}
	tokenString, err := generateToken(item.Username)
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "cannot_generate_token", "Cannot generate JWT", err)
		return
	}
	req.ReturnOK(w, r, loginResult{Token: tokenString}, 1)
}

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)
	tokenString, err := generateToken("admin")
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "cannot_generate_token", "Cannot generate JWT", err)
		return
	}
	req.ReturnOK(w, r, loginResult{Token: tokenString}, 1)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	req.ReturnOK(w, r, nil, 0)
}
