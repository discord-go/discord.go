module auditbot

go 1.26.4

replace github.com/discord-go/discord.go => ../../../..

	require github.com/discord-go/discord.go v0.0.0-00010101000000-000000000000

require (
	github.com/gorilla/websocket v1.5.3
	github.com/disgoorg/godave v0.3.0
	github.com/thomas-vilte/dave-go v0.5.1
	github.com/thomas-vilte/mls-go v1.6.0
	golang.org/x/crypto v0.54.0
	golang.org/x/sys v0.47.0
)
