package structs

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

type TransmitterStorage struct {
	Id              int
	Active          bool
	TransmitterType string
	URL             string
}
