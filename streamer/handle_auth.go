package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResult struct {
	Token       string `json:"token"`
	TokenExpiry int    `json:"tokenExpiry"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var item loginRequest
	req := PrepareReq(w, r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("cannot read body: %v", err)
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
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": item.Username,
			"exp":      time.Now().Add(time.Duration(config.Auth.TTL) * time.Second).Unix(),
		})

	tokenString, err := token.SignedString([]byte(config.Auth.JWTKey))
	if err != nil {
		log.Error("cannot generate token", "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "cannot_generate_token", "Cannot generate JWT", err)
		return
	}
	req.ReturnOK(w, r, loginResult{Token: tokenString, TokenExpiry: config.Auth.TTL}, 1)
}

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)
	username := "admin"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Duration(config.Auth.TTL) * time.Second).Unix(),
		})

	tokenString, err := token.SignedString([]byte(config.Auth.JWTKey))
	if err != nil {
		log.Error("cannot generate token", "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "cannot_generate_token", "Cannot generate JWT", err)
		return
	}
	req.ReturnOK(w, r, loginResult{Token: tokenString, TokenExpiry: config.Auth.TTL}, 1)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	req.ReturnOK(w, r, nil, 0)
}
