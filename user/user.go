package user

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"

	"github.com/CEKlopfenstein/gotify-repeater/relay"
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

func BuildInterface(basePath string, mux *gin.RouterGroup, relay *relay.Relay) {
	var cards = []card{}
	cards = append(cards, card{Title: "Discord Hook", Body: template.HTML("Hello")})
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
