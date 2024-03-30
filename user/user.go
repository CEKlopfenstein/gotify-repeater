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

//go:embed wrapper.html
var wrapper string

//go:embed htmx.min.js
var htmxMinJS string

//go:embed main.js
var mainJS string

type userPage struct {
	HtmxBasePath string
	Cards        []card
	MainJSPath   string
	pluginToken  string
}

type card struct {
	Title string
	Body  template.HTML
}

func buildConfigCard(config *structs.Config) template.HTML {
	tmpl, err := template.New("").Parse("<div><div>Server URL: {{.ServerURL}}</div><div>Client Token: {{.ClientToken}}</div><div>Discord Webhook: {{.DiscordWebHook}}</div></div>")
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
	var pageData = userPage{HtmxBasePath: "htmx.min.js", Cards: cards, MainJSPath: "main.js"}

	log.Println(basePath)
	log.Println(mux.BasePath())

	mux.GET("/", func(ctx *gin.Context) {
		log.Println(ctx.Request.Host)
		var clientKey = ctx.Request.Header.Get("X-Gotify-Key")
		if len(clientKey) == 0 {
			tmpl, err := template.New("").Parse(wrapper)
			if err != nil {
				log.Println(err)
				ctx.Done()
				return
			}
			err = tmpl.Execute(ctx.Writer, pageData)
			if err != nil {
				log.Println(err)
			}
			ctx.Done()
		} else {
			tmpl, err := template.New("").Parse(main)
			if err != nil {
				log.Println(err)
				ctx.Done()
				return
			}
			err = tmpl.Execute(ctx.Writer, pageData)
			if err != nil {
				log.Println(err)
			}
			ctx.Done()
		}

	})

	mux.GET("/"+pageData.HtmxBasePath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/javascript", []byte(htmxMinJS))
	})
	mux.GET("/"+pageData.MainJSPath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/javascript", []byte(mainJS))
	})

	mux.GET("/test", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html", []byte(pageData.pluginToken))
	})
}
