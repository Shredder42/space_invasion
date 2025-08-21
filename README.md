# Space Invasion

Space invasion (my take on the classic arcade game Space Invaders) is a real-time online multiplayer game where two players battle together against alien invaders. This repository contains both a Go game server and Go client code for the complete experience. 

The server runs on a local machine. HTTP requests are made to the server to create users and log in. Once logged in, the a connection is made via web socket between the client and server to pass game information back and forth. The server keeps track of game data such as asset location and collisions between bullets and enemies and shares it with the client while the client renders images and sends input to the server. Players must create an account and log in to connect to the game server. Once two players are connected, the game begins.


## Running the Space Invasion server

The Space Invasion server requires [Go](https://go.dev/doc/install) version 1.22 or higher and a postgres database

### Environment and Config

You will need a .env file that contains private environment variables necessary for the Space Invasion server to function.

* `DB_URL` - url string to connect to the postgres database
    It will look something like this `postgres://postgres:postgres@localhost:5432/space_invasion?sslmode=disable`
* `PLATFORM` - specifies what platform is accessing the server (e.g. dev)
* `SECRET` - an internal JWT used for authentication and authorization (Don't share this!!!)

You can load these variables into your main file with: 
```go
    godotenv.Load()
```
Then an example of loading the `dbURL`:
```go
    dbURL := os.Getenv("DB_URL")
    if dbURL == "" {
        log.Fatalf("DB_URL must be set")
    }
```
Then create an `apiConfig` in your main file like so:
```go
	apiCfg := apiConfig{
		jwtSecret:      jwtSecret,
		platform:       platform,
		db:             dbQueries,
	}
```

### Database
You will need to create the database in Postgres. Once you start the postgres server `sudo service postgresql start` on linux or `brew services start postgresql@15` on Mac, you can create the database using this command from the psql command line interface:
```sql
    CREATE DATABASE space_invasion;
```

Then you will need to run the database migrations in the sql/schema folder. To do this, you will need [Goose](https://github.com/pressly/goose), which can be installed using `go install github.com/pressly/goose/v3/cmd/goose@latest`
Then you can use a command like below to continue through the database migrations and set up the database.:
```bash
goose postgres "postgres://postgres:postgres@localhost:5432/space_invasion" up
```

### Start the server

From within the server directory, run `go run .` or `go build && ./server` to start the server.

## Running the Space Invasion client

The Space Invasion client requires [Go](https://go.dev/doc/install) version 1.22 or higher.

### Environment

The client relies on a .env file that contains the IP address of the server. This is private and is not shared on this github. The address can be obtained from the .env file the same way the server does it above.

In your main file:
```go
    godotenv.Load()
```
Then an example of loading the `serverAddress` IP address:
```go
    serverAddress := os.Getenv("GAME_SERVER")
    if serverAddress == "" {
        log.Fatalf("GAME_SERVER must be set")
    } 
```

### Run the client

While the server is running, the client can connect to the game. From within the client directory, run `go run .` or `go build && ./client` to join the game. You will be given prompts to create an account or log in if you already have an account. Once you are logged in, the client code will connect you to the game!

Playing the game is simple. Use the left and right arrow keys to move your ship and the spacebar to fire.


## Future steps
* Inject the IP address of the game server into the client at compile time. That way a .env file will not be required to connect to the game server
* Store player high scores in the data base when they disconnect
* Build more of the game, including things like:
    * new levels
    * different kinds of enemies
    * power-ups

