package transmitter

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
)

type LogTransmittor struct {
}

func (trans LogTransmittor) BuildTransmitterFunction() func(msg structs.GotifyMessageStruct, server server.Server) {
	return func(msg structs.GotifyMessageStruct, server server.Server) {
		log.Println(msg)
	}
}

func (trans LogTransmittor) HTMLCard() string {
	return "<div><h2>Log Transmitter</h2><div>\"Transmits\" the message to the Logs. Useful for Debugging.</div></div>"
}

type DiscordTransmitter struct {
	username string
	discord  string
}

type DiscordWebhookPayload struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

type DiscordHookInfo struct {
	Name string
}

func BuildDiscordTransmitter(discordHook string, name string) DiscordTransmitter {
	var transmitter = DiscordTransmitter{discord: discordHook}

	var hookInfo, err = transmitter.getHookInfo()
	if err != nil {
		transmitter.username = name
	} else {
		transmitter.username = hookInfo.Name
	}

	return transmitter
}

func (trans *DiscordTransmitter) getHookInfo() (DiscordHookInfo, error) {
	var hookInfo = DiscordHookInfo{}
	resp, err := http.Get(trans.discord)
	if err != nil {
		return hookInfo, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return hookInfo, err
	}

	err = json.Unmarshal(body, &hookInfo)
	if err != nil {
		return hookInfo, err
	}
	log.Println(hookInfo)
	return hookInfo, nil
}

func (trans DiscordTransmitter) BuildTransmitterFunction() func(msg structs.GotifyMessageStruct, server server.Server) {
	return func(msg structs.GotifyMessageStruct, server server.Server) {
		username := trans.username
		application, err := server.GetApplication(msg.Appid)
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

func (trans DiscordTransmitter) HTMLCard() string {
	var toReturn = `<div>
	<h2>Discord Webhook</h2>
	<div>
		<div>Default Username: ` + trans.username + `</div>
		<div>Webhook URL: ` + trans.discord + `</div>
	</div>
	</div>`

	return toReturn
}
