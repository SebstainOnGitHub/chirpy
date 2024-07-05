package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

const metricsTemplate = `
<html>
	<body>
    	<h1>Welcome, Chirpy Admin</h1>
    	<p>Chirpy has been visited {{.}} times!</p>
	</body>
</html>`

func handleGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusBadRequest, "Invalid request method")
		return
	}

	strId := r.PathValue("id")

	id, err := strconv.Atoi(strId)

	if err != nil {
		log.Fatal(err)
	}

	DB, err := newDB("./database.json")

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating database")
		return
	}

	chirpArr, err := DB.getAllChirps()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting chirps")
		return
	}

	if id > len(chirpArr) || id < 1 {
		respondWithError(w, http.StatusNotFound, "ID not found")
		return
	} else {
		respondWithJSON(w, http.StatusOK, chirpArr[id-1])
	}

}

func handleGetChirps(w http.ResponseWriter, r *http.Request) {
	DB, err := newDB("./database.json")
	if err != nil {
		log.Fatal(err)
		return
	}

	DB.loadDB()

	chirpArr, err := DB.getAllChirps()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting chirps from database")
		return
	}

	respondWithJSON(w, 200, chirpArr)
}

// For Posts
func handleChirpValidation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid request method")
		return
	}
	DB, err := newDB("./database.json")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating database")
		return
	}
	newChirp, err := DB.createChirp(r.Body)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp")
		return
	}

	if len(newChirp.Chirp) > 140 || len(newChirp.Chirp) == 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid message length")
		return
	}
	DB.appendDB(newChirp)

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
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.New("metrics").Parse(metricsTemplate)

	if err != nil {
		http.Error(w, "Error creating template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	if err := tmpl.Execute(w, cfg.fileserverHits); err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	apiCfg := &apiConfig{}

	mux := http.NewServeMux()

	mux.Handle("/app/*", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	mux.HandleFunc("GET /api/chirps", handleGetChirps)

	mux.HandleFunc("/admin/metrics", apiCfg.handleAdminMetrics)

	mux.HandleFunc("GET /api/metrics", apiCfg.handleMetrics)

	mux.HandleFunc("/api/reset", apiCfg.handleReset)

	mux.HandleFunc("GET /api/healthz", handleHealth)

	mux.HandleFunc("/api/chirps", handleChirpValidation)

	mux.HandleFunc("/api/chirps/{id}", handleGetSingleChirp)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	srv.ListenAndServe()
}
