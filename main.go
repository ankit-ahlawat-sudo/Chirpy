package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ankit-ahlawat-sudo/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type appConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
	platform string
	secret string
	polkaKey string
}

func main() {
	const filepathRoot = "."
	const port = "6969"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	platform:= os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	polkaKey:= os.Getenv("POLKA_KEY")

	secret:= os.Getenv("SECRET")

	dbQueries := database.New(db)

	apiCfg := appConfig{
		fileserverHits: atomic.Int32{},
		dbQueries: dbQueries,
		platform: platform,
		secret: secret,
		polkaKey: polkaKey,
	}

	mux := http.NewServeMux()

	handler:= http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerCountMetricsfunc)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetMetricsfunc)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUserAddition)
	mux.HandleFunc("POST /api/chirps", apiCfg.addChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirpsByCreateTime)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpsById)
	mux.HandleFunc("POST /api/login", apiCfg.handleUserLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.handleRevoke)
	mux.HandleFunc("PUT /api/users", apiCfg.handeEmailUpdate)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.upgradeToRed)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
