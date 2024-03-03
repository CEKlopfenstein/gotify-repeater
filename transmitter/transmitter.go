package transmitter

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/CEKlopfenstein/gotify-repeater/relay"
	"github.com/CEKlopfenstein/gotify-repeater/server"
)

type DiscordTransmitter struct {
	server   server.Server
	Username string
	discord  string
}

type DiscordWebhookPayload struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

func BuildDiscordTransmitter(server server.Server, discordHook string) DiscordTransmitter {
	return DiscordTransmitter{server: server, discord: discordHook}
}

func (trans *DiscordTransmitter) BuildTransmitterFunction() func(msg relay.GotifyMessageStruct) {
	return func(msg relay.GotifyMessageStruct) {
		username := trans.Username
		application, err := trans.server.GetApplication(msg.Appid)
		if err == nil {
			username = application.Name
		}

		var discordPayload = DiscordWebhookPayload{Username: username, Content: "# " + msg.Title + "\n\n" + msg.Message}

		discordBytePayload, err := json.Marshal(&discordPayload)
		if err != nil {
			log.Println("Failed To Build Discord Webhook Payload:", err.Error())
			return
		}
		resp, err := http.Post(trans.discord, "application/json", bytes.NewReader(discordBytePayload))
		if err != nil {
			log.Println("Failed to Send Discord Webhook:", err.Error())
			return
		} else if resp.StatusCode != http.StatusNoContent {
			log.Println("Discord Webhook returned response other than 204. Response:", resp.Status)
		}
	}
}
