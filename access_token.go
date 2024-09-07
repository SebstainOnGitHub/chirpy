package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func createRefreshToken(userID int) (DB_Refr_Token, error) {
	randArr := make([]byte, 32)
	_, err := rand.Read(randArr)
	if err != nil {
		return DB_Refr_Token{}, err
	}

	refrToken := DB_Refr_Token{
		ID: userID,
		Expiry_Time:   time.Now().Add(time.Duration(60) * 24 * time.Duration(time.Hour)),
		Refresh_Token: hex.EncodeToString(randArr),
	}
	return refrToken, nil
}
