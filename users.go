package main

import (
	"encoding/json"
	"net/http"
)

func(cfg *appConfig) handlerUserAddition(w http.ResponseWriter, r *http.Request){
	type requestBody struct {
		Email string `json:"email"`
	}

	decoder:= json.NewDecoder(r.Body)

	reqBody:= requestBody{}

	if err:= decoder.Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	email:= reqBody.Email

	user, err:= cfg.dbQueries.CreateUser(r.Context(), email)

	if err!=nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't add user", err)
		return
	}

	respondWithJSON(w, 201, User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	})

}