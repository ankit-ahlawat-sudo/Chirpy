package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ankit-ahlawat-sudo/Chirpy/internal/auth"
	"github.com/ankit-ahlawat-sudo/Chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        	uuid.UUID `json:"id"`
	CreatedAt 	time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func(cfg *appConfig) handlerUserAddition(w http.ResponseWriter, r *http.Request){

	type paremeters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder:= json.NewDecoder(r.Body)

	reqBody:= paremeters{}

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
		IsChirpyRed: user.IsChirpyRed.Bool,
	})

}

func(cfg *appConfig) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		// ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type response struct {
		User 
		Token  string `json:"token"` 
		RefreshToken string `json:"refresh_token"`
	}

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

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh JWT", err)
		return
	}
 
	refreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh JWT", err)
		return
	}

	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: user.ID,
		ExpiresAt: time.Now().Add(60*24*time.Hour),
	})


	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	respondWithJSON(w, 200, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			IsChirpyRed: user.IsChirpyRed.Bool,
		},
		Token: token,
		RefreshToken: refreshToken,
	})
}

func(cfg *appConfig) handeEmailUpdate(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	accesstoken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find token", err)
		return
	}

	decoder:= json.NewDecoder(r.Body);
	req:= request{}

	if err:= decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	userId, err:= auth.ValidateJWT(accesstoken, cfg.secret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't Validation JWT", err)
		return
	}
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash Passowrd", err)
		return
	}

	user, err:= cfg.dbQueries.UpdateEmailPassword(r.Context(), database.UpdateEmailPasswordParams{
		Email: req.Email,
		HashedPassword: hashedPassword,
		ID: userId,
	})

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate User", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			IsChirpyRed: user.IsChirpyRed.Bool,
	})

}

func(cfg *appConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	user, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}

	accessToken, err:= auth.MakeJWT(user.ID, cfg.secret, time.Hour)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})

}

func(cfg *appConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	_, err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func(cfg *appConfig) upgradeToRed(w http.ResponseWriter, r *http.Request) {
	type data struct {
		User_id string `json:"user_id"`
	}
	type request struct {
		Event string `json:"event"`
		Data data `json:"data"`
	}

	decoder:= json.NewDecoder(r.Body)

	apiKey, err := auth.GetAPIKey(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid ID", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "invalid ID", err)
		return
	}

	var requestData request

	if err:= decoder.Decode(&requestData); err!= nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if requestData.Event!= "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	validUuid, err:= uuid.Parse(requestData.Data.User_id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID", err)
		return
	}

	err= cfg.dbQueries.UpgradeToRedById(r.Context(), validUuid)

	if err != nil {
		respondWithError(w,http.StatusNotFound, "id not found to upgrade", nil)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}

