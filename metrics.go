package main

import (
	"fmt"
	"net/http"
	"os"
)

func (cfg *appConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
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