package main

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]chirp `json:"chirps"`
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
		_, err := os.Create(db.path)
		if err != nil {
			return err
		}
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
	if dbstruct.Chirps == nil {
		return DBStructure{Chirps: map[int]chirp{}}, nil
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

func (db *DB) createID() (int, error) {
	chirpArr, err := db.getAllChirps()
	if err != nil {
		return -1, err
	}
	return len(chirpArr) + 1, nil
}

func (db *DB) createChirp(data io.ReadCloser) (chirp, error) {
	dec := json.NewDecoder(data)

	id, err := db.createID()

	if err != nil {
		return chirp{}, err
	}

	newChirp := chirp{
		ID:    id,
		Chirp: "",
	}

	dec.Decode(&newChirp)

	newChirp.filterForProfane()

	return newChirp, nil
}

func (db *DB) appendDB(chirp chirp) error {
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
