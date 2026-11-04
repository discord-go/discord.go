module auditbot

go 1.22.0

replace github.com/discord-go/discord.go => ../../../..

replace github.com/discord-go/discord.go/dave-go => ../../../../dave-go

replace github.com/discord-go/discord.go/dave-go/godave => ../../../../dave-go/godave

replace github.com/discord-go/discord.go/dave-go/mls-go => ../../../../dave-go/mls-go

require github.com/discord-go/discord.go v0.0.0-00010101000000-000000000000

require (
	github.com/discord-go/discord.go/dave-go v0.1.0 // indirect
	github.com/discord-go/discord.go/dave-go/godave v0.1.0 // indirect
	github.com/discord-go/discord.go/dave-go/mls-go v0.1.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/time v0.10.0 // indirect
)
