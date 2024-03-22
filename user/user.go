package user

import (
	"bytes"
	_ "embed"
	"html/template"
	"log"
	"net/http"

	"github.com/CEKlopfenstein/gotify-repeater/relay"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/gin-gonic/gin"
)

//go:embed main.html
var main string

//go:embed htmx.min.js
var htmxMin string

type userPage struct {
	HtmxBasePath string
	Cards        []card
}

type card struct {
	Title string
	Body  template.HTML
}

func buildConfigCard(config *structs.Config) template.HTML {
	tmpl, err := template.New("").Parse("<div><div>{{.ServerURL}}</div><div>{{.ClientToken}}</div><div>{{.DiscordWebHook}}</div></div>")
	if err != nil {
		return template.HTML("Error: " + err.Error())
	}
	var doc bytes.Buffer
	err = tmpl.Execute(&doc, config)
	if err != nil {
		return template.HTML("Error: " + err.Error())
	}
	return template.HTML(doc.String())
}

func BuildInterface(basePath string, mux *gin.RouterGroup, relay *relay.Relay, hookConfig *structs.Config) {
	var cards = []card{}
	cards = append(cards, card{Title: "Discord Hook", Body: buildConfigCard(hookConfig)})
	var pageData = userPage{HtmxBasePath: "htmx.min.js", Cards: cards}

	log.Println(basePath)

	mux.GET("/", func(ctx *gin.Context) {
		tmpl, err := template.New("").Parse(main)
		if err != nil {
			log.Println(err)
			ctx.Done()
			return
		}
		log.Println("Test")
		err = tmpl.Execute(ctx.Writer, pageData)
		if err != nil {
			log.Println(err)
		}
		ctx.Done()
	})
	mux.GET("/"+pageData.HtmxBasePath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/javascript", []byte(htmxMin))
	})
}
