package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Shredder42/space_invasion/shared"
	"github.com/gorilla/websocket"
)

type Credentials struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type User struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"user_name"`
	Token     string    `json:"token"`
}

func getCredentials() (string, string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("(C)reate account or (L)ogin")
	option, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading option: %v", err)
	}

	fmt.Println("Enter username: ")
	userName, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading user name: %v", err)
	}

	fmt.Println("Enter password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading password: %v", err)
	}

	option = strings.Trim(option, "\n")
	userName = strings.Trim(userName, "\n")
	password = strings.Trim(password, "\n")

	return strings.ToLower(option), userName, password

}

func createAccount(userName, password, path string) {
	credentials := Credentials{
		UserName: userName,
		Password: password,
	}

	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		log.Fatal("error marshaling credentials: %v", err)
		return
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(credentialsJSON))
	if err != nil {
		log.Fatal("error creating request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal("error making request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		log.Fatal("account not created")
		return
	}

	log.Printf(res.Status)
	log.Printf("Account created for user: %s", userName)

}

func loginUser(username, password, path string) string {
	credentials := Credentials{
		UserName: username,
		Password: password,
	}

	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		log.Fatal("error marshaling credentials: %v", err)
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(credentialsJSON))
	if err != nil {
		log.Fatal("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal("error making request: %v", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("error reading response: %v", err)
	}

	var user User
	err = json.Unmarshal(data, &user)
	if err != nil {
		log.Fatal("error unmarshaling JSON: %v", err)
	}

	log.Println(res.Status)
	if res.StatusCode != http.StatusOK {
		log.Println(string(data))
		log.Fatal("couldn't log in")
	}

	log.Printf("user %s logged in", user.Username)

	return user.Token

}

func (g *Game) connectToGameServer(userName, serverAddress string) {
	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", g.token))
	header.Set("Username", userName)

	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:8080/ws", serverAddress), header)
	if err != nil {
		log.Printf("Connection failed: %v", err)
		return
	}

	g.conn = conn
	g.connected = true
	log.Println("Connected to server!")
}

func (g *Game) listenForGameServerMessages() {
	for {
		var message shared.ServerMessage
		err := g.conn.ReadJSON(&message)
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			return
		}

		switch message.Type {
		case "player_id":
			log.Printf(message.PlayerID)
			g.myPlayerID = message.PlayerID
		case "game_state":
			g.updateClientPlayers(message.GameState)
			g.updateClientBullets(message.GameState)
			g.updateClientEnemies(message.GameState)
		}
	}
}
