package user

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/CEKlopfenstein/gotify-repeater/relay"
	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/CEKlopfenstein/gotify-repeater/storage"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	transmitter "github.com/CEKlopfenstein/gotify-repeater/transmitters"
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

//go:embed transmitter-select.html
var transmitterSelect string

//go:embed bootstrap.min.css
var bootstrap string

type userPage struct {
	HtmxBasePath string
	Cards        []card
	MainJSPath   string
	Bootstrap    string
}

type card struct {
	Title string
	Body  template.HTML
}

func BuildInterface(basePath string, mux *gin.RouterGroup, relay *relay.Relay, hookConfig *structs.Config, c storage.Storage, hostname string) {
	var cards = []card{}
	var pageData = userPage{HtmxBasePath: "htmx.min.js", Cards: cards, MainJSPath: "main.js", Bootstrap: "bootstrap.min.css"}

	mux.GET("/"+pageData.HtmxBasePath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/javascript", []byte(htmxMinJS))
	})
	mux.GET("/"+pageData.MainJSPath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/javascript", []byte(mainJS))
	})
	mux.GET("/"+pageData.Bootstrap, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/css", []byte(bootstrap))
	})

	mux.GET("/", func(ctx *gin.Context) {
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
			var server = relay.GetServer()
			var failed = server.CheckToken(clientKey)
			if failed != nil {
				log.Println(failed)
				ctx.Data(http.StatusOK, "text/html", []byte("<h2>Unauthorized token. Redirecting to main page.</h2><script>window.location = '/';</script>"))
				ctx.Done()
				return
			}
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
		}

	})

	internalServer := server.SetupServer(hostname, "")
	mux.Use(func(ctx *gin.Context) {
		var clientKey = ctx.Request.Header.Get("X-Gotify-Key")
		if len(clientKey) == 0 {
			ctx.Data(http.StatusUnauthorized, "text/html", []byte("X-Gotify-Key Missing"))
			ctx.Done()
			return
		}

		var failed = internalServer.UpdateToken(clientKey)
		if failed != nil {
			log.Println(failed)
			ctx.Data(http.StatusUnauthorized, "application/json", []byte(failed.Error()))
			ctx.Done()
			return
		}
		ctx.Set("token", clientKey)
		ctx.Next()
	})

	mux.GET("/getLoginToken", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html", []byte(ctx.GetString("token")))
	})

	mux.GET("/transmitters", func(ctx *gin.Context) {
		var transmitters = relay.GetTransmitters()
		var cards = ""
		for key := range transmitters {
			cards += transmitters[key].HTMLCard(key)
		}
		ctx.Data(http.StatusOK, "text/html", []byte(cards))
	})

	transmitterGroup := mux.Group("/transmitter/:transmitterID", func(ctx *gin.Context) {
		var transmitters = relay.GetTransmitters()
		var id = ctx.Param("transmitterID")
		var intId, _ = strconv.Atoi(id)

		log.Println("IDs", id, intId)

		var transmitter = transmitters[intId]
		if transmitter == nil {
			ctx.Data(http.StatusNotFound, "text/html", []byte("Invalid ID"))
			return
		}

		ctx.Set("transID", intId)
		ctx.Next()
	})

	transmitterGroup.GET("/", func(ctx *gin.Context) {
		var transmitter transmitter.Transmitter
		var id = ctx.GetInt("transID")

		var transmitters = relay.GetTransmitters()
		transmitter = transmitters[id]

		ctx.Data(http.StatusOK, "text/html", []byte(transmitter.HTMLCard(id)))
	})

	transmitterGroup.DELETE("/", func(ctx *gin.Context) {
		var id = ctx.GetInt("transID")
		relay.RemoveTransmitter(id)
		ctx.Data(http.StatusOK, "text/html", []byte(""))
	})

	transmitterGroup.PUT("/status", func(ctx *gin.Context) {
		var id = ctx.GetInt("transID")

		var status = ctx.PostForm("active")
		var boolStatus = status == "on"

		relay.SetTransmitterStatus(id, boolStatus)

		ctx.Data(http.StatusOK, "text/html", []byte(fmt.Sprintf(`<input class="form-check-input" hx-put="transmitter/%d/status" hx-trigger="click changed" type="checkbox" value="%s" name="active">`, id, status)))
	})

	mux.GET("/transmitter-options", func(ctx *gin.Context) {
		tmpl, _ := template.New("").Parse(transmitterSelect)
		var buffer bytes.Buffer
		type internal struct {
			Types map[string]string
		}
		var types = internal{Types: transmitter.TransmitterTypes}
		tmpl.Execute(&buffer, types)
		ctx.Data(http.StatusOK, "text/html", buffer.Bytes())
	})

	mux.PUT("/transmitter-select", func(ctx *gin.Context) {
		var transmitterType = ctx.PostForm("transmitter")
		var function = transmitter.TransmitterCreationPages[transmitterType]
		if function != nil {
			ctx.Data(http.StatusOK, "text/html", function(transmitterType))
			return
		}

		ctx.Data(http.StatusBadRequest, "text/html", []byte("<div>Invalid Transmitter Type Selected</div>"))
	})

	mux.POST("/transmitter-select", func(ctx *gin.Context) {
		var transmitterType = ctx.PostForm("transmitter")
		var function = transmitter.TransmitterCreationPostHandler[transmitterType]
		if function != nil {
			var data = function(transmitterType, ctx, c.AddTransmitter, c.GetCurrentTransmitterNextID())
			relay.ReloadTransmitters()
			ctx.Data(http.StatusOK, "text/html", data)
			return
		}
		ctx.Data(http.StatusBadRequest, "text/html", []byte("<div>Invalid Transmitter Type Selected</div>"))
	})

	mux.GET("/defaultToken", func(ctx *gin.Context) {
		var token = c.GetClientToken()
		if len(token) == 0 {
			ctx.Data(http.StatusOK, "text/html", []byte(`<div hx-target="this" hx-swap="outerHTML">
			<div>No Token Set. Select an option below to set one.</div>
			<button class="btn btn-secondary m-1" hx-put="defaultToken" hx-vals='js:{"token":localStorage.getItem("gotify-login-key")}'>Use Current Client Token</button><button class="btn btn-secondary m-1" hx-put="defaultToken" hx-vals='{"token":"new"}'>Create Custom Client Token</button>
			</div>`))
		} else {
			ctx.Data(http.StatusOK, "text/html", []byte(`<div hx-target="this" hx-swap="outerHTML">
			<div>Current Default Token: `+token+`</div>
			<div>Use Options Below to Change Token</div>
			<button class="btn btn-secondary m-1" hx-put="defaultToken" hx-vals='js:{"token":localStorage.getItem("gotify-login-key")}'>Use Current Client Token</button><button class="btn btn-secondary m-1" hx-put="defaultToken" hx-vals='{"token":"new"}'>Create Custom Client Token</button>
			</div>`))
		}

	})

	mux.PUT("/defaultToken", func(ctx *gin.Context) {
		var headerToken = ctx.GetString("token")
		var token = ctx.PostForm("token")

		if token == "new" {
			currentToken := c.GetClientToken()
			if internalServer.CheckToken(currentToken) == nil {
				client := internalServer.FindClientFromToken(currentToken)
				if len(client.Token) != 0 && client.Token != headerToken && client.Name == "Relay Client" {
					internalServer.DeleteClient(client.Id)
				}
			}
			newClient, err := internalServer.CreateClient("Relay Client")
			if err != nil {
				log.Println(err)
				ctx.Redirect(303, "defaultToken")
				return
			}
			token = newClient.Token
		}

		relay.UpdateToken(token)

		ctx.Redirect(303, "defaultToken")
	})
}
