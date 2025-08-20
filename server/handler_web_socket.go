package main

import (
	"log"
	"net/http"

	"github.com/Shredder42/space_invasion/server/internal/auth"
	"github.com/Shredder42/space_invasion/shared"
)

func (gs *GameServer) handlerWebSocket(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting bearer token", err)
		return
	}

	_, err = auth.ValidateJWT(token, gs.apiConfig.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid web token", err)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade to websocket failed: %v", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected!")

	playerName := gs.addNewPlayer(conn)
	log.Printf("Player %s connected", playerName)

	for {
		var action shared.PlayerAction

		err := conn.ReadJSON(&action)
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			break
		}

		if action.Type == "move" {
			gs.players[action.ID].MovePlayer(action.Direction)
		}
		if action.Type == "shoot" {
			gs.shoot(action.ID)
		}
	}

}
