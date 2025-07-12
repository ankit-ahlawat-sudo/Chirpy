package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request){
	type parameter struct {
		Body string `json:"body" `
	}
	type validRes struct {
		Valid bool `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}
	decoder:= json.NewDecoder(r.Body)
	paras:= parameter{}

	err:= decoder.Decode(&paras)
	if err!= nil{
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if len(paras.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", err)
		return
	}

	words:= strings.Split(paras.Body, " ")

	for i, word := range words {
		lowerWord := strings.ToLower(word)
		if lowerWord == "kerfuffle" || lowerWord == "sharbert" || lowerWord == "fornax" {
			words[i] = "****"
		}
	}


	respondWithJSON(w, http.StatusOK, validRes{
		Valid: true,
		CleanedBody: strings.Join(words, " "),
	})
}
