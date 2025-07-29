module github.com/Shredder42/space_invasion/client

go 1.24.2

require github.com/hajimehoshi/ebiten/v2 v2.8.8

require (
	github.com/Shredder42/space_invasion/shared v0.0.0
	github.com/ebitengine/gomobile v0.0.0-20240911145611-4856209ac325 // indirect
	github.com/ebitengine/hideconsole v1.0.0 // indirect
	github.com/ebitengine/purego v0.8.0 // indirect
	github.com/gorilla/websocket v1.5.3
	github.com/jezek/xgb v1.1.1 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

replace github.com/Shredder42/space_invasion/shared => ../shared
