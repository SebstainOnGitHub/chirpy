package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/joho/godotenv"
)

const metricsTemplate = `
<html>
	<body>
    	<h1>Welcome, Chirpy Admin</h1>
    	<p>Chirpy has been visited {{.}} times!</p>
	</body>
</html>`

const pathToDB = "./database.json"

func handleRevokeAccessToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusBadRequest, "invalid request method")
		return
	}

	hdr := r.Header.Get("Authorization")

	if hdr == "" {
		respondWithError(w, http.StatusBadRequest, "header(s) not present")
		return
	}

	DB, err := newDB(pathToDB)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating database")
		return
	}

	err = DB.findAndDeleteRefrToken(hdr)
	
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

// Handles refresh endpoint
func (apicfg *apiConfig) handleVerifyAccessToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusBadRequest, "invalid request method")
		return
	}

	db, err := newDB(pathToDB)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating database")
		return
	}

	hdr := r.Header.Get("Authorization")

	if hdr == "" {
		respondWithError(w, http.StatusBadRequest, "header(s) not present")
		return
	}

	userID, err := db.validateRefreshToken(hdr)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error validating refresh token")
		return
	}

	user, err := db.getUsrByID(userID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error finding user")
		return
	}

	newJWT, err := apicfg.createJWT(user)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating a new refresh token")
		return
	}

	respondWithJSON(w, http.StatusOK, newJWT)
}

// Handles updating user info with a jwt, nothing else
func (apicfg *apiConfig) handleVerifyJWT(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusBadRequest, "invalid request method")
		return
	}

	if r.Header.Get("Authorization") == "" {
		respondWithError(w, http.StatusBadRequest, "header(s) not present")
		return
	}

	userID, err := apicfg.validateJWT(r.Header.Get("Authorization"))

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error validating token")
		return
	}

	DB, err := newDB(pathToDB)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating database")
		return
	}

	userDetailsInRequest, err := DB.createTempUser(r.Body)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	userInDB, err := DB.updateUser(&userDetailsInRequest, userID)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, userInDB.omitPassword())
}

// Handles creating a JWT and a refresh token to login in future. The refr token is just used to make a new JWT to log in again.
func (apicfg *apiConfig) handleCreateJWT(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusBadRequest, "invalid request method")
		return
	}

	DB, err := newDB(pathToDB)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating database")
		return
	}

	createdUser, err := DB.validatePotential(r.Body)

	if err != nil{
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	jwtResp, err := apicfg.createJWTWithResponse(createdUser)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating JWT")
		return
	}

	respondWithJSON(w, 200, jwtResp)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusBadRequest, "invalid request method")
		return
	}

	DB, err := newDB(pathToDB)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating database")
		return
	}

	createdUser, err := DB.createUser(r.Body)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = DB.appendDBUser(createdUser)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, createdUser.omitPassword())
}

func handleGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusBadRequest, "invalid request method")
		return
	}

	strId := r.PathValue("id")

	id, err := strconv.Atoi(strId)

	if err != nil {
		log.Fatal(err)
	}

	DB, err := newDB(pathToDB)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating database")
		return
	}

	chirpArr, err := DB.getAllChirps()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error getting chirps")
		return
	}

	if id > len(chirpArr) || id < 1 {
		respondWithError(w, http.StatusNotFound, "id not found")
	} else {
		respondWithJSON(w, http.StatusOK, chirpArr[id-1])
	}

}

func handleGetChirps(w http.ResponseWriter, r *http.Request) {
	DB, err := newDB(pathToDB)
	if err != nil {
		log.Fatal(err)
		return
	}

	DB.loadDB()

	chirpArr, err := DB.getAllChirps()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error getting chirps from database")
		return
	}

	respondWithJSON(w, 200, chirpArr)
}

// For Posts
func handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "invalid request method")
		return
	}
	DB, err := newDB(pathToDB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating database")
		return
	}
	newChirp, err := DB.createChirp(r.Body)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating chirp")
		return
	}

	if len(newChirp.Chirp) > 140 || len(newChirp.Chirp) == 0 {
		respondWithError(w, http.StatusBadRequest, "invalid message length")
		return
	}

	DB.appendDBChirp(newChirp)

	respondWithJSON(w, 201, newChirp)
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits)))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(200)
}

func (cfg *apiConfig) handleAdminMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.New("metrics").Parse(metricsTemplate)

	if err != nil {
		http.Error(w, "error creating template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	if err := tmpl.Execute(w, cfg.fileserverHits); err != nil {
		http.Error(w, "error executing template", http.StatusInternalServerError)
		return
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	godotenv.Load()

	jwtSecret := os.Getenv("JWT_SECRET")

	apiCfg := &apiConfig{jwtSecret: jwtSecret}

	mux := http.NewServeMux()

	mux.Handle("/app/*", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	mux.HandleFunc("/admin/metrics", apiCfg.handleAdminMetrics)

	mux.HandleFunc("/api/metrics", apiCfg.handleMetrics)

	mux.HandleFunc("/api/reset", apiCfg.handleReset)

	mux.HandleFunc("/api/healthz", handleHealth)

	mux.HandleFunc("/api/login", apiCfg.handleCreateJWT)

	mux.HandleFunc("/api/refresh", apiCfg.handleVerifyAccessToken)

	mux.HandleFunc("/api/users", handleCreateUser)

	mux.HandleFunc("PUT /api/users", apiCfg.handleVerifyJWT)

	mux.HandleFunc("/api/revoke", handleRevokeAccessToken)

	mux.HandleFunc("/api/chirps", handleCreateChirp)

	mux.HandleFunc("GET /api/chirps", handleGetChirps)

	mux.HandleFunc("/api/chirps/{id}", handleGetSingleChirp)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	srv.ListenAndServe()
}
