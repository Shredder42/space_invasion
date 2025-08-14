package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserName  string    `json:"user_name"`
}

func (cfg *apiConfig) handlerCreateUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		UserName string `json:"user_name"`
	}

	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
	}

	user, err := cfg.db.CreateUser(r.Context(), params.UserName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			UserName:  user.UserName,
		},
	})

	log.Printf("Created user %s at %v\n", user.UserName, user.CreatedAt)

}
