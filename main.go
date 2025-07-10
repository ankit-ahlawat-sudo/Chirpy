package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	// "golang.org/x/tools/go/cfg"
)

type appConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "6969"

	apiCfg := new(appConfig)

	mux := http.NewServeMux()

	handler:= http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	// mux.HandleFunc("GET /api/metrics", apiCfg.handlerCountMetricsfunc)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerCountMetricsfunc)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetMetricsfunc)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *appConfig) handlerCountMetricsfunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	count := cfg.fileserverHits.Load()
	htmlBytes, err := os.ReadFile("./metricCount.html")
	if err != nil {
		http.Error(w, "Could not read metricCount file", http.StatusInternalServerError)
		return
	}
	htmlString := string(htmlBytes)

	html:= fmt.Sprintf(htmlString, count)

	w.Write([]byte(html))
	
}

func(cfg *appConfig) handlerResetMetricsfunc(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func (cfg *appConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
