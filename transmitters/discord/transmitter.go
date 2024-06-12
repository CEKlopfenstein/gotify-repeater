package discordTransmitter

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/CEKlopfenstein/gotify-repeater/gotify_api"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/gin-gonic/gin"
)

type DiscordTransmitter struct {
	username      string
	discord       string
	status        bool
	transmitCount int
}

type DiscordWebhookPayload struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

type DiscordHookInfo struct {
	Name string
}

func Build(discordHook string, name string, status bool, count int) DiscordTransmitter {
	var transmitter = DiscordTransmitter{discord: discordHook}

	var hookInfo, err = transmitter.getHookInfo()
	if err != nil {
		transmitter.username = name
	} else {
		transmitter.username = hookInfo.Name
	}

	transmitter.SetStatus(status)

	transmitter.transmitCount = count

	return transmitter
}

//go:embed new.html
var transmitterCreationForm string

var globalLogger *log.Logger

func SetGlobalLogger(logger *log.Logger) {
	globalLogger = logger
}

type transmitterCreationFormData struct {
	Type string
	HTMX template.HTML
}

func NewTransmitterForm(transmitterType string) []byte {
	templ, err := template.New("").Parse(transmitterCreationForm)

	if err != nil {
		globalLogger.Println(err)
	}

	var buffer = bytes.Buffer{}

	err = templ.Execute(&buffer, transmitterCreationFormData{Type: transmitterType})

	if err != nil {
		globalLogger.Println(err)
	}

	return buffer.Bytes()
}

func CreateTransmitterFromForm(transmitterType string, ctx *gin.Context, storeFunction func(transmitter structs.TransmitterStorage) int, id int) []byte {
	var transmitter = Build(ctx.PostForm("discord-url"), fmt.Sprintf("Transmitter %d", id), true, 0)
	storeFunction(transmitter.GetStorageValue(id))
	templ, err := template.New("").Parse(transmitterCreationForm)

	if err != nil {
		globalLogger.Println(err)
	}

	var buffer = bytes.Buffer{}

	err = templ.Execute(&buffer, transmitterCreationFormData{Type: transmitterType, HTMX: template.HTML(`<span hx-swap="beforebegin" hx-target="closest #newTransmitters" hx-get="transmitter/` + fmt.Sprint(id) + `" hx-trigger="load once"></span>`)})

	if err != nil {
		globalLogger.Println(err)
	}

	return buffer.Bytes()
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
	return hookInfo, nil
}

func (trans *DiscordTransmitter) Transmit(msg structs.GotifyMessageStruct, server gotify_api.GotifyApi) {
	username := trans.username
	application, err := server.GetApplication(msg.Appid)
	if err == nil {
		username = application.Name
	}

	var discordPayload = DiscordWebhookPayload{Username: username, Content: "# " + msg.Title + "\n\n" + msg.Message}

	discordBytePayload, err := json.Marshal(&discordPayload)
	if err != nil {
		globalLogger.Println("Failed To Build Discord Webhook Payload:", err.Error())
		return
	}
	resp, err := http.Post(trans.discord, "application/json", bytes.NewReader(discordBytePayload))
	if err != nil {
		globalLogger.Println("Failed to Send Discord Webhook:", err.Error())
		return
	} else if resp.StatusCode != http.StatusNoContent {
		globalLogger.Println("Discord Webhook returned response other than 204. Response:", resp.Status)
	} else {
		trans.transmitCount++
	}
}

//go:embed card.html
var card string

func (trans DiscordTransmitter) HTMLCard(id int) string {

	template, err := template.New("").Parse(card)
	if err != nil {
		globalLogger.Println(err)
		return err.Error()
	}
	writer := bytes.Buffer{}
	type temp struct {
		Username   string
		DiscordURL string
		ID         int
		Status     string
	}
	data := temp{ID: id, Username: trans.username, DiscordURL: trans.discord}

	if trans.Active() {
		data.Status = "checked"
	} else {
		data.Status = ""
	}

	err = template.Execute(&writer, data)
	if err != nil {
		globalLogger.Println(err)
		return err.Error()
	}

	return writer.String()
}

func (trans DiscordTransmitter) GetStorageValue(id int) structs.TransmitterStorage {
	return structs.TransmitterStorage{Id: id, URLorTOKEN: trans.discord, TransmitterType: "discord", Active: trans.Active(), TransmitCount: trans.GetTransmitCount()}
}

func (trans DiscordTransmitter) Active() bool {
	return trans.status
}

func (trans *DiscordTransmitter) SetStatus(active bool) {
	trans.status = active
}

func (trans *DiscordTransmitter) GetTransmitCount() int {
	return trans.transmitCount
}
