package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ankit-ahlawat-sudo/Chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct{
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *appConfig) addChirp(w http.ResponseWriter, r *http.Request){
	type ChirpMessage struct{
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	decoder:= json.NewDecoder(r.Body)

	chirpMessage:= ChirpMessage{}

	if err:= decoder.Decode(&chirpMessage); err!= nil {
		respondWithError(w, 500, "Not able to decode", err)
		return
	}

	const maxChirpLength = 140
	if len(chirpMessage.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(chirpMessage.Body, badWords)

	chirp, err:= cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: cleaned,
		UserID: chirpMessage.UserId,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Not able to add the chirp", err)
		return
	}

	respondWithJSON(w, 201, Chirp{
		ID: chirp.ID,
		Body: chirp.Body,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserID: chirp.UserID,
	})
	
}

func(cfg *appConfig) getChirpsByCreateTime(w http.ResponseWriter, r *http.Request) {
	chirps, err:= cfg.dbQueries.GetChirpsDortedByCreateTime(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Not able to get the chirps", err)
		return
	}

	apiChirps := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		apiChirps[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	respondWithJSON(w, 200, apiChirps)
}

func(cfg *appConfig) getChirpsById(w http.ResponseWriter, r *http.Request) {
	chirpIDstring := r.PathValue("chirpID")
	chirpID, err:= uuid.Parse(chirpIDstring)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID", err)
		return
	}
	chirp, err:= cfg.dbQueries.GetChirpById(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not able to fetch the chirp", err)
		return
	}

	respondWithJSON(w, 200, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})

}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}