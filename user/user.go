package user

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/CEKlopfenstein/gotify-repeater/relay"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/gin-gonic/gin"
	"github.com/gotify/plugin-api"
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
	log.Println(config)
	err = tmpl.Execute(&doc, config)
	if err != nil {
		return template.HTML("Error: " + err.Error())
	}
	return template.HTML(doc.String())
}

func BuildInterface(basePath string, mux *gin.RouterGroup, relay *relay.Relay, hookConfig *structs.Config, c plugin.StorageHandler) {
	var cards = []card{}
	cards = append(cards, card{Title: "Discord Hook", Body: buildConfigCard(hookConfig)})
	var pageData = userPage{HtmxBasePath: "htmx.min.js", Cards: cards, MainJSPath: "main.js"}

	log.Println(basePath)
	log.Println(mux.BasePath())

	mux.GET("/"+pageData.HtmxBasePath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/javascript", []byte(htmxMinJS))
	})
	mux.GET("/"+pageData.MainJSPath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/javascript", []byte(mainJS))
	})

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

	mux.Use(func(ctx *gin.Context) {
		var clientKey = ctx.Request.Header.Get("X-Gotify-Key")
		if len(clientKey) == 0 {
			ctx.Data(http.StatusUnauthorized, "text/html", []byte("X-Gotify-Key Missing"))
			ctx.Done()
			return
		}
		var server = relay.GetServer()
		var failed = server.CheckToken(clientKey)
		if failed != nil {
			ctx.Data(http.StatusUnauthorized, "application/json", []byte(failed.Error()))
			ctx.Done()
			return
		}

		ctx.Next()
	})

	mux.GET("/getLoginToken", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html", []byte(ctx.Request.Header.Get("X-Gotify-Key")))
	})

	mux.GET("/transmitters", func(ctx *gin.Context) {
		var transmitters = relay.GetTransmitters()
		var cards = ""
		for key := range transmitters {
			cards += transmitters[key].HTMLCard()
		}
		ctx.Data(http.StatusOK, "text/html", []byte(cards))
	})

	mux.GET("/test", func(ctx *gin.Context) {
		log.Println(ctx.Request.Header.Get("X-Gotify-Key"))
		ctx.Data(http.StatusOK, "text/html", []byte(pageData.pluginToken))
	})

	type contact struct {
		FirstName string
		LastName  string
		Email     string
	}

	mux.GET("/contact", func(ctx *gin.Context) {
		storageBytes, _ := c.Load()
		var contactInfo = contact{}
		if len(storageBytes) == 0 {
			contactInfo.FirstName = ""
			contactInfo.LastName = ""
			contactInfo.Email = ""
			storageBytes, _ = json.Marshal(contactInfo)
			c.Save(storageBytes)
		} else {
			json.Unmarshal(storageBytes, &contactInfo)
		}
		ctx.Data(http.StatusOK, "text/html", []byte(`<div hx-target="this" hx-swap="outerHTML">
        <div><label>First Name</label>: `+contactInfo.FirstName+`</div>
        <div><label>Last Name</label>: `+contactInfo.LastName+`</div>
        <div><label>Email</label>: `+contactInfo.Email+`</div>
        <button hx-get="edit" class="btn btn-primary">
        Click To Edit
        </button>
    </div>`))
	})

	mux.GET("/edit", func(ctx *gin.Context) {
		storageBytes, _ := c.Load()
		var contactInfo = contact{}
		if len(storageBytes) == 0 {
			contactInfo.FirstName = ""
			contactInfo.LastName = ""
			contactInfo.Email = ""
			storageBytes, _ = json.Marshal(contactInfo)
			c.Save(storageBytes)
		} else {
			json.Unmarshal(storageBytes, &contactInfo)
		}
		ctx.Data(http.StatusOK, "text/html", []byte(`<form hx-put="contact" hx-target="this" hx-swap="outerHTML">
		<div>
		  <label>First Name</label>
		  <input type="text" name="firstName" value="`+contactInfo.FirstName+`">
		</div>
		<div class="form-group">
		  <label>Last Name</label>
		  <input type="text" name="lastName" value="`+contactInfo.LastName+`">
		</div>
		<div class="form-group">
		  <label>Email Address</label>
		  <input type="email" name="email" value="`+contactInfo.Email+`">
		</div>
		<button class="btn">Submit</button>
		<button class="btn" hx-get="contact">Cancel</button>
	  </form>`))
	})

	mux.PUT("/contact", func(ctx *gin.Context) {
		var contactInfo = contact{}
		contactInfo.FirstName = ctx.PostForm("firstName")
		contactInfo.LastName = ctx.PostForm("lastName")
		contactInfo.Email = ctx.PostForm("email")
		storageBytes, _ := json.Marshal(contactInfo)
		c.Save(storageBytes)
		ctx.Redirect(303, "contact")
	})
}
