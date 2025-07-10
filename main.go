package main

import (
	"log"
	"net/http"
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
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerCountMetricsfunc)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetMetricsfunc)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
