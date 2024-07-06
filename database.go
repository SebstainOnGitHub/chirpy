package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]chirp `json:"chirps"`
	Users  map[int]user  `json:"users"`
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
		return DBStructure{Chirps: map[int]chirp{}, Users: map[int]user{}}, nil
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

func (db *DB) getAllUsers() ([]user, error) {
	userArr := []user{}
	dbstruct, err := db.loadDB()
	if err != nil {
		return []user{}, err
	}
	for _, val := range dbstruct.Users {
		userArr = append(userArr, val)
	}
	return userArr, nil
}

func (db *DB) createChirpID() (int, error) {
	chirpArr, err := db.getAllChirps()
	if err != nil {
		return -1, err
	}
	return len(chirpArr) + 1, nil
}

func (db *DB) createUserID() (int, error) {
	userArr, err := db.getAllUsers()
	if err != nil {
		return -1, err
	}
	return len(userArr) + 1, nil
}

func (db *DB) createChirp(data io.ReadCloser) (chirp, error) {
	dec := json.NewDecoder(data)

	id, err := db.createChirpID()

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

func (db *DB) createTempUser(body io.ReadCloser) (user, error) {
	dec := json.NewDecoder(body)
	id, err := db.createUserID()
	if err != nil {
		return user{}, nil
	}
	newUser := user{
		ID: id - 1,
	}

	dec.Decode(&newUser)

	return newUser, nil
}

func (db *DB) createUser(body io.ReadCloser) (user, error) {
	dec := json.NewDecoder(body)
	id, err := db.createUserID()
	if err != nil {
		return user{}, nil
	}
	newUser := user{
		ID: id,
	}

	dec.Decode(&newUser)

	if _, exists := db.getByEmail(newUser.Email); exists {
		return user{}, errors.New("email already exists")
	}

	newUser.Password, err = bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)

	if err != nil {
		return user{}, err
	}

	return newUser, nil
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

func (db *DB) getByEmail(email string) (user, bool) {
	dbstruct, err := db.loadDB()

	if err != nil {
		log.Fatal(err)
		return user{}, false
	}
	for _, val := range dbstruct.Users {
		if val.Email == email {
			return val, true
		}
	}
	return user{}, false
}

func (db *DB) validateLogin(logged user) (int, error) {
	foundUser, exists := db.getByEmail(logged.Email)
	if !exists {
		return http.StatusNotFound, errors.New("email not found")
	}

	if err := bcrypt.CompareHashAndPassword(foundUser.Password, []byte(logged.Password)); err != nil {
		return http.StatusUnauthorized, errors.New("invalid password")
	}

	return http.StatusOK, nil
}
