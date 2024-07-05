package main

import (
	"encoding/json"
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
	if err == os.ErrNotExist {
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

	if err = json.Unmarshal(data, &dbstruct); err != nil {
		return DBStructure{}, err
	}
	return dbstruct, nil
}

func (db *DB) writeDB(dbstruct DBStructure) error {
	db.mux.RLock()
	defer db.mux.RUnlock()
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

func (db *DB) createChirp(body string) (chirp, error) {
	id, err := db.createID()

	if err != nil {
		return chirp{}, err
	}

	newChirp := chirp{
		ID:    id,
		Chirp: body,
	}

	newChirp.filterForProfane()

	dbstruct, err := db.loadDB()

	if err != nil {
		return chirp{}, err
	}

	dbstruct.Chirps[newChirp.ID] = newChirp
	
	db.writeDB(dbstruct)

	return newChirp, nil
}

func (db *DB) writeChirp(chirp chirp) error {
	data, err := json.Marshal(chirp)

	if err != nil {
		return err
	}

	f, err := os.OpenFile(db.path, os.O_APPEND|os.O_WRONLY, 0644) 

	if err != nil {
		return err
	}

	f.Write(data)

	f.Close()

	return nil
}