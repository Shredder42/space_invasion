module github.com/Shredder42/space_invasion/server

go 1.24.2

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	golang.org/x/crypto v0.41.0 // indirect
)

require github.com/Shredder42/space_invasion/shared v0.0.0

replace github.com/Shredder42/space_invasion/shared => ../shared
