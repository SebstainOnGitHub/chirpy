package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
)


const metricsTemplate = `
<html>
	<body>
    	<h1>Welcome, Chirpy Admin</h1>
    	<p>Chirpy has been visited {{.}} times!</p>
	</body>
</html>`


func handleGetChirps(w http.ResponseWriter, r *http.Request) {
	DB, err := newDB("./database.json")
	if err != nil {
		log.Fatal(err)
	}

	DB.loadDB()

	chirpArr, err := DB.getAllChirps()
	
	if err != nil {
		respondWithError(w, 500, err.Error())
	}

	respondWithJSON(w, 200, chirpArr)
}


//For Posts
func handleChirpValidation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	dec := json.NewDecoder(r.Body)
	newChirp := chirp{}
	err := dec.Decode(&newChirp)

	if err != nil {
		http.Error(w, "Something went wrong", 500)
		return
	}

	if len(newChirp.Chirp) > 140 || len(newChirp.Chirp) == 0 {
		w.WriteHeader(400)
		return
	}

	DB, err := newDB("./database.json")
	if err != nil {
		log.Fatal(err)
	}

	DB.loadDB()
	
	DB.writeChirp(newChirp)

	respondWithJSON(w, 200, newChirp)
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

	mux.HandleFunc("GET /api/chirps", handleGetChirps )

	mux.HandleFunc("/admin/metrics", apiCfg.handleAdminMetrics)

	mux.HandleFunc("GET /api/metrics", apiCfg.handleMetrics)

	mux.HandleFunc("/api/reset", apiCfg.handleReset)

	mux.HandleFunc("GET /api/healthz", handleHealth)

	mux.HandleFunc("/api/chirps", handleChirpValidation)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	srv.ListenAndServe()
}
