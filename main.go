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
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	dbQueries := database.New(db)
	
	const filepathRoot = "."
	const port = "6969"

	apiCfg := appConfig{
		fileserverHits: atomic.Int32{},
		dbQueries: dbQueries,
	}

	mux := http.NewServeMux()

	handler:= http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerCountMetricsfunc)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetMetricsfunc)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
