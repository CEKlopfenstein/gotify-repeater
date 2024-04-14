package transmitters

import (
	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	discordTransmitter "github.com/CEKlopfenstein/gotify-repeater/transmitters/discord"
	logTransmitter "github.com/CEKlopfenstein/gotify-repeater/transmitters/log"
	"github.com/gin-gonic/gin"
)

// Transmitter Interface that represents structs capable of transmitting from the relay.
type Transmitter interface {
	// Returns HTML representation of the Transmitter
	HTMLCard(int) string
	// Dehydrates the Transmitter regardless of type into a Struct that can be safely stored for later.
	GetStorageValue(int) structs.TransmitterStorage
	// Transmit using this transmitter
	Transmit(msg structs.GotifyMessageStruct, server server.Server)
	// Gets a boolean to indicate if it's active
	Active() bool
	SetStatus(bool)
}

type TransmitterType struct {
	Name                string
	Full_Name           string
	CreationPage        (func(string) []byte)
	CreationPostHandler (func(string, *gin.Context, func(transmitter structs.TransmitterStorage) int, int) []byte)
}

var Types = map[string]TransmitterType{
	"log": {
		Name:                "log",
		Full_Name:           "Log Transmitter",
		CreationPage:        logTransmitter.NewTransmitterForm,
		CreationPostHandler: logTransmitter.CreateTransmitterFromForm},
	"discord": {
		Name:                "discord",
		Full_Name:           "Discord Web Hook",
		CreationPage:        discordTransmitter.NewTransmitterForm,
		CreationPostHandler: discordTransmitter.CreateTransmitterFromForm}}

func RehydrateTransmitter(stored structs.TransmitterStorage) Transmitter {
	if stored.TransmitterType == "discord" {
		trans := discordTransmitter.BuildDiscordTransmitter(stored.URL, "Default Name", stored.Active)
		return &trans
	} else if stored.TransmitterType == "log" {
		trans := logTransmitter.Build(stored.Active)
		return &trans
	}
	return &logTransmitter.LogTransmittor{}
}
