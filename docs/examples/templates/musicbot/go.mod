module musicbot

go 1.22.0

require (
	github.com/discord-go/discord.go v0.0.0
	github.com/gorilla/websocket v1.5.3
	layeh.com/gopus v0.0.0-20210501142526-1ee02d434e32
)

require (
	github.com/discord-go/discord.go/dave-go v0.1.0 // indirect
	github.com/discord-go/discord.go/dave-go/godave v0.1.0 // indirect
	github.com/discord-go/discord.go/dave-go/mls-go v0.1.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/time v0.10.0 // indirect
)

replace github.com/discord-go/discord.go => ../../../..

replace github.com/discord-go/discord.go/dave-go => ../../../../dave-go

replace github.com/discord-go/discord.go/dave-go/godave => ../../../../dave-go/godave

replace github.com/discord-go/discord.go/dave-go/mls-go => ../../../../dave-go/mls-go
