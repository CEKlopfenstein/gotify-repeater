package structs

// Contains Structs that I need to be able to have intialized in other packages.
// Without causing circular dependancies.

type Config struct {
	DiscordWebHook string
	ClientToken    string
	ServerURL      string
}
