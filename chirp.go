package main

import (
	"strings"
)

type chirp struct {
	ID    int    `json:id`
	Chirp string `json:"body"`
}

func (chirp *chirp) filterForProfane() string {
	val := chirp.Chirp

	splitVal := strings.Split(strings.ToLower(val), " ")

	returnStr := make([]string, len(splitVal))

	for i, val := range splitVal {
		if val == "kerfuffle" || val == "sharbert" || val == "fornax" {
			returnStr[i] = "****"
		} else {
			returnStr[i] = val
		}
	}
	strChirp := strings.Join(returnStr, " ")
	chirp.Chirp = strChirp
	return strChirp
}
