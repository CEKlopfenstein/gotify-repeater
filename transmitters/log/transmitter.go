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

func HTMLNewForm(transmitterType string) []byte {
	var test = `<form hx-post="transmitter-select" hx-target="this" hx-swap="outerHTML">
		<input type="hidden" name="transmitter" value="` + transmitterType + `">
		<div>
		  No farther values required. Click Submit to create.
		</div>
		<button class="btn btn-primary">Submit</button>
	  </form>`
	return []byte(test)
}

func HTMLCreate(transmitterType string, ctx *gin.Context, storeFunction func(transmitter structs.TransmitterStorage) int, id int) []byte {
	var transmitter = Build(true)
	storeFunction(transmitter.GetStorageValue(id))
	var test = `<form hx-post="transmitter-select" hx-target="this" hx-swap="outerHTML">
	<input type="hidden" name="transmitter" value="` + transmitterType + `" hx-swap="beforebegin" hx-target="closest .newTransmitters" hx-get="transmitter/` + fmt.Sprint(id) + `" hx-trigger="load once">
	<div>
	  No farther values required. Click Submit to create.
	</div>
	<button class="btn btn-primary">Submit</button>
  </form>`
	return []byte(test)
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
