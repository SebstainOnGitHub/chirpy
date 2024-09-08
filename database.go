package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps         map[int]chirp   `json:"chirps"`
	Users          map[int]user    `json:"users"`
	Refresh_Tokens []DB_Refr_Token `json:"refresh_tokens"`
}

type DB_Refr_Token struct {
	ID            int       `json:"id"`
	Expiry_Time   time.Time `json:"expiry_time"`
	Refresh_Token string    `json:"refresh_token"`
}

func newDB(path string) (*DB, error) {
	newDB := DB{path, &sync.RWMutex{}}
	err := newDB.ensureDB()
	if err != nil {
		return &DB{}, err
	}
	return &newDB, nil
}

func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)
	if os.IsNotExist(err) {
		f, err := os.Create(db.path)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	data, err := os.ReadFile(db.path)

	if err != nil {
		return DBStructure{}, err
	}

	dbstruct := DBStructure{}

	err = json.Unmarshal(data, &dbstruct)

	//If file empty
	if dbstruct.Chirps == nil && dbstruct.Users == nil {
		return DBStructure{Chirps: map[int]chirp{}, Users: map[int]user{}, Refresh_Tokens: []DB_Refr_Token{}}, nil
	}

	if err != nil {
		return DBStructure{}, err
	}

	return dbstruct, nil
}

func (db *DB) writeDB(dbstruct DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	data, err := json.Marshal(dbstruct)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, data, os.FileMode(0644))

	if err != nil {
		return err
	}
	return nil
}

func (db *DB) getAllChirps() ([]chirp, error) {
	chirpArr := []chirp{}
	dbstruct, err := db.loadDB()
	if err != nil {
		return []chirp{}, err
	}
	for _, val := range dbstruct.Chirps {
		chirpArr = append(chirpArr, val)
	}
	return chirpArr, nil
}

func (db *DB) getAllRefreshTokens() ([]DB_Refr_Token, error) {
	refrTokenArr := []DB_Refr_Token{}
	dbstruct, err := db.loadDB()
	if err != nil {
		return []DB_Refr_Token{}, err
	}
	return append(refrTokenArr, dbstruct.Refresh_Tokens...), nil
}

func (db *DB) newChirpID() (int, error) {
	dbstruct, err := db.loadDB()
	if err != nil {
		return -1, nil
	}
	return len(dbstruct.Chirps) + 1, nil
}

func (db *DB) newUserID() (int, error) {
	dbstruct, err := db.loadDB()
	if err != nil {
		return -1, nil
	}
	return len(dbstruct.Users) + 1, nil
}

func (db *DB) createChirpStruct(data io.ReadCloser) (chirp, error) {
	dec := json.NewDecoder(data)

	id, err := db.newChirpID()

	if err != nil {
		return chirp{}, err
	}

	newChirp := chirp{
		ID: id,
	}

	err = dec.Decode(&newChirp)

	if err != nil {
		return chirp{}, nil
	}

	newChirp.filterForProfane()

	return newChirp, nil
}

func (db *DB) createChirp(data io.ReadCloser, userID int) (chirp, error) {
	newChirp, err := db.createChirpStruct(data)

	if err != nil {
		return chirp{}, err
	}

	newChirp.Author_ID = userID

	err = db.appendDBChirp(newChirp)

	if err != nil {
		return chirp{}, err
	}

	return newChirp, nil
}

func (db *DB) validatePotential(body io.ReadCloser) (user, error) {
	newUser := jsonUser{}

	dec := json.NewDecoder(body)

	err := dec.Decode(&newUser)

	if err != nil {
		return user{}, errors.New("error decoding request")
	}

	potUser, exists := db.getByEmail(newUser.Email)

	if !exists || bcrypt.CompareHashAndPassword(potUser.Password, []byte(newUser.Password)) != nil {
		return user{}, errors.New("invalid login details, please try again")
	}

	return potUser, nil
}

func (db *DB) createTempUser(body io.ReadCloser) (user, error) {
	defer body.Close()

	newUser := jsonUser{}

	id, err := db.newUserID()

	if err != nil {
		return user{}, err
	}

	finalUser := user{
		ID: id,
	}

	dec := json.NewDecoder(body)

	err = dec.Decode(&newUser)

	if err != nil {
		return user{}, err
	}

	finalUser.Password, err = bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)

	finalUser.Email = newUser.Email

	if err != nil {
		return user{}, errors.New("error creating password")
	}

	return finalUser, nil
}

func (db *DB) createUser(body io.ReadCloser) (user, error) {
	defer body.Close()

	newUser := jsonUser{}

	id, err := db.newUserID()

	if err != nil {
		return user{}, err
	}

	finalUser := user{
		ID:            id,
		Is_Chirpy_Red: false,
	}

	dec := json.NewDecoder(body)

	err = dec.Decode(&newUser)

	if err != nil {
		return user{}, err
	}

	finalUser.Password, err = bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)

	if err != nil {
		return user{}, errors.New("error creating password")
	}

	if _, exists := db.getByEmail(newUser.Email); exists {
		return user{}, errors.New("email already exists")
	}

	finalUser.Email = newUser.Email

	return finalUser, nil
}

func (db *DB) appendDBChirp(chirp chirp) error {
	dbStruct, err := db.loadDB()
	if err != nil {
		return err
	}

	dbStruct.Chirps[chirp.ID] = chirp

	err = db.writeDB(dbStruct)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) appendDBUser(user user) error {
	dbStruct, err := db.loadDB()
	if err != nil {
		return err
	}

	dbStruct.Users[user.ID] = user

	err = db.writeDB(dbStruct)

	if err != nil {
		return err
	}

	return nil
}

func (db *DB) appendDBRefrToken(refrToken DB_Refr_Token) error {
	dbstruct, err := db.loadDB()

	if err != nil {
		return err
	}

	dbstruct.Refresh_Tokens = append(dbstruct.Refresh_Tokens, refrToken)

	err = db.writeDB(dbstruct)

	if err != nil {
		return err
	}

	return nil
}

func (db *DB) updateUser(updatedUser *user, userID int) (user, error) {
	dbstruct, err := db.loadDB()

	if err != nil {
		return user{}, err
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(updatedUser.Password), bcrypt.DefaultCost)

	if err != nil {
		return user{}, err
	}

	updatedUser.Password = hashedPass

	if _, ok := dbstruct.Users[userID]; !ok {
		return user{}, errors.New("user not found")
	}

	usrNotPtr := user{
		ID:       userID,
		Email:    updatedUser.Email,
		Password: updatedUser.Password,
		Is_Chirpy_Red: updatedUser.Is_Chirpy_Red,
	}

	dbstruct.Users[userID] = usrNotPtr

	db.writeDB(dbstruct)

	return usrNotPtr, nil
}

func (db *DB) makeAndStoreRefreshToken(userID int) (DB_Refr_Token, error) {
	newRefrTokenString, err := createRefreshToken(userID)

	if err != nil {
		return DB_Refr_Token{}, err
	}

	newRefrToken := DB_Refr_Token{
		ID:            userID,
		Expiry_Time:   time.Now().Add(1 * time.Hour),
		Refresh_Token: newRefrTokenString.Refresh_Token,
	}

	err = db.appendDBRefrToken(newRefrToken)

	if err != nil {
		return DB_Refr_Token{}, err
	}

	return newRefrToken, nil
}

func (db *DB) removeRefrToken(tokenstr string) error {
	dbstruct, err := db.loadDB()

	if err != nil {
		return err
	}

	token, err := db.getRefrByToken(tokenstr)

	if err != nil {
		return err
	}

	ommitedTokenArr := dbstruct.Refresh_Tokens[:indexOfRefr(token, dbstruct.Refresh_Tokens)]
	ommitedTokenArr = append(ommitedTokenArr, dbstruct.Refresh_Tokens[indexOfRefr(token, dbstruct.Refresh_Tokens)+1:]...)

	dbstruct.Refresh_Tokens = ommitedTokenArr

	db.writeDB(dbstruct)

	return nil
}

func (db *DB) findAndDeleteRefrToken(header string) error {
	if len(header) < 7 {
		return errors.New("no header found")
	}

	refr_token_string := header[7:]

	fmt.Println(refr_token_string)

	Refr_TokenArr, err := db.getAllRefreshTokens()

	if err != nil {
		return err
	}

	for _, val := range Refr_TokenArr {
		if val.Refresh_Token == refr_token_string && time.Now().Before(val.Expiry_Time) {
			fmt.Println("token found")
			db.removeRefrToken(val.Refresh_Token)
			return nil
		} else if time.Now().Before(val.Expiry_Time) {
			db.removeRefrToken(val.Refresh_Token)
		}
	}

	return errors.New("token not found")
}

func (db *DB) validateRefreshToken(refr_token_string_with_bearer string) (int, error) {
	if len(refr_token_string_with_bearer) < 7 {
		return -1, errors.New("no header found")
	}

	refr_token_string := refr_token_string_with_bearer[7:]

	dbstruct, err := db.loadDB()

	if err != nil {
		return -1, err
	}

	for _, val := range dbstruct.Refresh_Tokens {
		if val.Refresh_Token == refr_token_string && time.Now().Before(val.Expiry_Time) {
			return val.ID, nil
		}
	}

	return -1, errors.New("refresh token not found")
}

func (db *DB) getUsrByID(id int) (user, error) {
	dbstruct, err := db.loadDB()
	if err != nil {
		return user{}, err
	}
	usr, ok := dbstruct.Users[id]
	if !ok {
		return user{}, errors.New("user not found")
	}
	return usr, nil
}

func (db *DB) getRefrByToken(tokenstr string) (DB_Refr_Token, error) {
	dbstruct, err := db.loadDB()
	if err != nil {
		return DB_Refr_Token{}, err
	}

	for _, val := range dbstruct.Refresh_Tokens {
		if tokenstr == val.Refresh_Token {
			return val, nil
		}
	}

	return DB_Refr_Token{}, errors.New("token not found")
}

func indexOfRefr(token DB_Refr_Token, refrArr []DB_Refr_Token) int {
	for i, val := range refrArr {
		if token.Refresh_Token == val.Refresh_Token {
			return i
		}
	}
	return -1
}

func (db *DB) getByEmail(email string) (user, bool) {
	dbstruct, err := db.loadDB()

	if err != nil {
		return user{}, false
	}

	for _, val := range dbstruct.Users {
		if val.Email == email {
			return val, true
		}
	}

	return user{}, false
}

func (db *DB) upgradeUser(id int) error {
	dbstruct, err := db.loadDB()
	
	if err != nil {
		return err
	}

	if dbUser, ok := dbstruct.Users[id]; ok{
		dbUser.Is_Chirpy_Red = true

		dbstruct.Users[id] = dbUser
		
		db.writeDB(dbstruct)
		return nil
	}

	return errors.New("user does not exist")
}

func (db *DB) decodeWebhook(data io.ReadCloser) error {
	webhook := webhookBody{}

	var err error

	json.NewDecoder(data).Decode(&webhook)

	fmt.Println(webhook)

	defer data.Close()

	if webhook.Event == "user.upgraded" {
		err = db.upgradeUser(webhook.Data.ID)
	}

	return err
}
