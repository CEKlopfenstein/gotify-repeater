package pushbulletTransmitter

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/CEKlopfenstein/gotify-repeater/gotify_api"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/gin-gonic/gin"
)

type PushBulletTransmitter struct {
	url           string
	AccessToken   string
	DefaultTitle  string
	transmitCount int
	status        bool
}

type PushBulletPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Type  string `json:"type"`
}

func Build(accesstoken string, name string, status bool, count int) PushBulletTransmitter {
	var transmitter = PushBulletTransmitter{url: "https://api.pushbullet.com/v2/pushes", AccessToken: accesstoken, DefaultTitle: name, transmitCount: count, status: status}
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
	var transmitter = Build(ctx.PostForm("pushbullet-token"), fmt.Sprintf("Transmitter %d", id), true, 0)

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

func (trans *PushBulletTransmitter) Transmit(msg structs.GotifyMessageStruct, server gotify_api.GotifyApi) {
	var pushBulletPayload PushBulletPayload
	pushBulletPayload.Type = "note"
	pushBulletPayload.Title = trans.DefaultTitle

	// Attempt to get title
	application, err := server.GetApplication(msg.Appid)
	if err == nil {
		pushBulletPayload.Title = application.Name
	}

	pushBulletPayload.Body = msg.Title + "\n" + msg.Message

	pushbulletBytePayload, err := json.Marshal(&pushBulletPayload)
	if err != nil {
		globalLogger.Println("Failed To Build Pushbullet Payload:", err.Error())
		return
	}

	client := http.Client{}
	req, err := http.NewRequest("POST", trans.url, bytes.NewReader(pushbulletBytePayload))

	if err != nil {
		globalLogger.Println("Failed To Build Pushbullet Request:", err.Error())
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Access-Token", trans.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		globalLogger.Println("Failed to Send Pushbullet:", err.Error())
		return
	} else if resp.StatusCode != http.StatusOK {
		globalLogger.Println("Pushbullet returned response other than 200. Response:", resp.Status)
	} else {
		trans.transmitCount++
	}
}

//go:embed card.html
var card string

func (trans PushBulletTransmitter) HTMLCard(id int) string {

	template, err := template.New("").Parse(card)
	if err != nil {
		globalLogger.Println(err)
		return err.Error()
	}
	writer := bytes.Buffer{}
	type temp struct {
		Title  string
		Token  string
		ID     int
		Status string
	}
	data := temp{ID: id, Title: trans.DefaultTitle, Token: trans.AccessToken}

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

func (trans PushBulletTransmitter) GetStorageValue(id int) structs.TransmitterStorage {
	return structs.TransmitterStorage{Id: id, URLorTOKEN: trans.AccessToken, TransmitterType: "pushbullet", Active: trans.Active(), TransmitCount: trans.GetTransmitCount()}
}

func (trans PushBulletTransmitter) Active() bool {
	return trans.status
}

func (trans *PushBulletTransmitter) SetStatus(active bool) {
	trans.status = active
}

func (trans *PushBulletTransmitter) GetTransmitCount() int {
	return trans.transmitCount
}
