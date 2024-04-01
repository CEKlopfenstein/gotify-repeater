package structs

import "github.com/CEKlopfenstein/gotify-repeater/server"

// Contains Structs that I need to be able to have intialized in other packages.
// Without causing circular dependancies.

type Config struct {
	DiscordWebHook string
	ClientToken    string
	ServerURL      string
}

type GotifyMessageStruct struct {
	Appid    int
	Date     string
	Id       int
	Message  string
	Title    string
	Priority int
}

type Transmitter interface {
	BuildTransmitterFunction() func(msg GotifyMessageStruct, server server.Server)
	HTMLCard() string
}
