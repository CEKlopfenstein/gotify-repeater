package transmitters

import (
	"log"

	"github.com/CEKlopfenstein/gotify-repeater/gotify_api"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	discordTransmitter "github.com/CEKlopfenstein/gotify-repeater/transmitters/discord"
	logTransmitter "github.com/CEKlopfenstein/gotify-repeater/transmitters/log"
	pushbulletTransmitter "github.com/CEKlopfenstein/gotify-repeater/transmitters/pushbullet"
	"github.com/gin-gonic/gin"
)

// Transmitter Interface that represents structs capable of transmitting from the relay.
type Transmitter interface {
	// Returns HTML representation of the Transmitter
	HTMLCard(int) string
	// Dehydrates the Transmitter regardless of type into a Struct that can be safely stored for later.
	GetStorageValue(int) structs.TransmitterStorage
	// Transmit using this transmitter
	Transmit(msg structs.GotifyMessageStruct, server gotify_api.GotifyApi)
	// Gets a boolean to indicate if it's active
	Active() bool
	SetStatus(bool)
	// Gets the number of times the transmitter has been fired. -1 Returned if not implemented
	GetTransmitCount() int
}

type TransmitterType struct {
	Name                string
	Full_Name           string
	CreationPage        (func(string) []byte)
	CreationPostHandler (func(string, *gin.Context, func(transmitter structs.TransmitterStorage) int, int) []byte)
	SetGlobalLogger     (func(*log.Logger))
}

var Types = map[string]TransmitterType{
	"log": {
		Name:                "log",
		Full_Name:           "Log Transmitter",
		CreationPage:        logTransmitter.NewTransmitterForm,
		CreationPostHandler: logTransmitter.CreateTransmitterFromForm,
		SetGlobalLogger:     logTransmitter.SetGlobalLogger},
	"discord": {
		Name:                "discord",
		Full_Name:           "Discord Web Hook",
		CreationPage:        discordTransmitter.NewTransmitterForm,
		CreationPostHandler: discordTransmitter.CreateTransmitterFromForm,
		SetGlobalLogger:     discordTransmitter.SetGlobalLogger},
	"pushbullet": {
		Name:                "pushbullet",
		Full_Name:           "Pushbullet",
		CreationPage:        pushbulletTransmitter.NewTransmitterForm,
		CreationPostHandler: pushbulletTransmitter.CreateTransmitterFromForm,
		SetGlobalLogger:     pushbulletTransmitter.SetGlobalLogger,
	}}

func RehydrateTransmitter(stored structs.TransmitterStorage) Transmitter {
	if stored.TransmitterType == "discord" {
		trans := discordTransmitter.Build(stored.URLorTOKEN, "Default Name", stored.Active, stored.TransmitCount)
		return &trans
	} else if stored.TransmitterType == "log" {
		trans := logTransmitter.Build(stored.Active, stored.TransmitCount)
		return &trans
	} else if stored.TransmitterType == "pushbullet" {
		trans := pushbulletTransmitter.Build(stored.URLorTOKEN, "Default Name", stored.Active, stored.TransmitCount)
		return &trans
	}
	return &logTransmitter.LogTransmittor{}
}
