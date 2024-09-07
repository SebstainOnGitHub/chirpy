package main

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtResponse struct {
	ID            int    `json:"id"`
	Email         string `json:"email"`
	Token         string `json:"token"`
	Refresh_Token string `json:"refresh_token"`
}

type jwtOnlyToken struct {
	Token string `json:"token"`
}

func (apicfg *apiConfig) createJWT(r user) (jwtOnlyToken, error) {
	claims := jwt.MapClaims{
		"iss": "chirpy",
		"iat": jwt.NewNumericDate(time.Now()),
		"exp": jwt.NewNumericDate(time.Now().Add(time.Duration(1) * time.Hour)),
		"sub": strconv.Itoa(r.ID),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err := jwtToken.SignedString([]byte(apicfg.jwtSecret))

	if err != nil {
		return jwtOnlyToken{}, err
	}

	respToken := jwtOnlyToken{Token: token}

	return respToken, nil
}

func (apicfg *apiConfig) createJWTWithResponse(r user) (jwtResponse, error) {
	claims := jwt.MapClaims{
		"iss": "chirpy",
		"iat": jwt.NewNumericDate(time.Now()),
		"exp": jwt.NewNumericDate(time.Now().Add(time.Duration(1) * time.Hour)),
		"sub": strconv.Itoa(r.ID),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err := jwtToken.SignedString([]byte(apicfg.jwtSecret))

	if err != nil {
		return jwtResponse{}, err
	}

	DB, err := newDB(pathToDB)

	if err != nil {
		return jwtResponse{}, nil
	}

	dbRefrToken, err := DB.makeAndStoreRefreshToken(r.ID)

	if err != nil {
		return jwtResponse{}, nil
	}

	return jwtResponse{
		ID:            r.ID,
		Email:         r.Email,
		Token:         token,
		Refresh_Token: dbRefrToken.Refresh_Token,
	}, nil
}

func (apicfg *apiConfig) validateJWT(header string) (string, error) {
	//"Bearer " needs to be stripped from the header
	if len(header) < 7 {
		return "", errors.New("no header found")
	}

	tokenString := header[7:]

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(apicfg.jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	userID, err := token.Claims.GetSubject()

	if err != nil {
		return "", err
	}

	return userID, nil
}