module github.com/discord-go/discord.go/dave-go

go 1.22

require (
	github.com/discord-go/discord.go/dave-go/godave v0.1.0
	github.com/discord-go/discord.go/dave-go/mls-go v0.1.0
)

require (
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

replace github.com/discord-go/discord.go/dave-go/godave => ./godave

replace github.com/discord-go/discord.go/dave-go/mls-go => ./mls-go
