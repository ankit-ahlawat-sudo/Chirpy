package main

import (
	"encoding/json"
	"net/http"

	"github.com/ankit-ahlawat-sudo/Chirpy/internal/auth"
	"github.com/ankit-ahlawat-sudo/Chirpy/internal/database"
)

type requestBody struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func(cfg *appConfig) handlerUserAddition(w http.ResponseWriter, r *http.Request){

	decoder:= json.NewDecoder(r.Body)

	reqBody:= requestBody{}

	if err:= decoder.Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	email:= reqBody.Email
	password, err:= auth.HashPassword(reqBody.Password)
	if err!= nil{
		respondWithError(w, http.StatusInternalServerError, "Couldn't encode password", err)
		return
	}

	user, err:= cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email: email,
		HashedPassword: password,
	})

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

func(cfg *appConfig) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	decoder:= json.NewDecoder(r.Body)

	reqBody:= requestBody{}

	if err:= decoder.Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	email:= reqBody.Email

	user,err:= cfg.dbQueries.GetUserByEmail(r.Context(), email)

	if err!=nil {
		respondWithError(w, 401, "Unauthorized", err)
		return
	}

	err= auth.CheckPasswordHash(reqBody.Password, user.HashedPassword)
	if err!= nil{
		respondWithError(w, 401, "Unauthorized", err)
		return
	}

	respondWithJSON(w, 200, User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	})
}