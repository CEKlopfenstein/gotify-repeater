package logTransmitter

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"

	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/gin-gonic/gin"
)

type LogTransmittor struct {
	status bool
}

func (trans LogTransmittor) Transmit(msg structs.GotifyMessageStruct, server server.Server) {
	log.Println("LogTransmittor, MSG:", msg.Message, "Priority:", msg.Priority, "Raw:", msg)
}

//go:embed card.html
var card string

func Build(status bool) LogTransmittor {
	var transmitter = LogTransmittor{}
	transmitter.SetStatus(status)
	return transmitter
}

//go:embed new.html
var transmitterCreationForm string

type transmitterCreationFormData struct {
	Type string
	HTMX template.HTML
}

func HTMLNewForm(transmitterType string) []byte {
	templ, err := template.New("").Parse(transmitterCreationForm)

	if err != nil {
		log.Println(err)
	}

	var buffer = bytes.Buffer{}

	err = templ.Execute(&buffer, transmitterCreationFormData{Type: transmitterType})

	if err != nil {
		log.Println(err)
	}

	return buffer.Bytes()
}

func HTMLCreate(transmitterType string, ctx *gin.Context, storeFunction func(transmitter structs.TransmitterStorage) int, id int) []byte {
	var transmitter = Build(true)
	storeFunction(transmitter.GetStorageValue(id))
	templ, err := template.New("").Parse(transmitterCreationForm)

	if err != nil {
		log.Println(err)
	}

	var buffer = bytes.Buffer{}

	err = templ.Execute(&buffer, transmitterCreationFormData{Type: transmitterType, HTMX: template.HTML(`<span hx-swap="beforebegin" hx-target="closest #newTransmitters" hx-get="transmitter/` + fmt.Sprint(id) + `" hx-trigger="load once"></span>`)})

	if err != nil {
		log.Println(err)
	}

	return buffer.Bytes()
}

func (trans LogTransmittor) HTMLCard(id int) string {
	template, err := template.New("").Parse(card)
	if err != nil {
		log.Println(err)
		return err.Error()
	}
	writer := bytes.Buffer{}
	type temp struct {
		ID     int
		Status string
	}
	data := temp{ID: id}
	if trans.Active() {
		data.Status = "checked"
	} else {
		data.Status = ""
	}

	err = template.Execute(&writer, data)
	if err != nil {
		log.Println(err)
		return err.Error()
	}

	return writer.String()
}

func (trans LogTransmittor) GetStorageValue(id int) structs.TransmitterStorage {
	return structs.TransmitterStorage{Id: id, TransmitterType: "log", Active: trans.Active()}
}

func (trans LogTransmittor) Active() bool {
	return trans.status
}

func (trans *LogTransmittor) SetStatus(active bool) {
	trans.status = active
}
