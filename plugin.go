package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gotify/plugin-api"
)

// GetGotifyPluginInfo returns gotify plugin info.
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		ModulePath:  "github.com/gotify/plugin-template",
		Version:     "1.0.0",
		Author:      "CEKlopfenstein",
		Description: "An example plugin with travis-ci building",
		Name:        "cekwebhooks",
	}
}

// MyPlugin is the gotify plugin instance.
type MyPlugin struct {
	userCtx     plugin.UserContext
	msgHandler  plugin.MessageHandler
	serverURL   string
	clientToken string
}

// Enable enables the plugin.
func (c *MyPlugin) Enable() error {
	go func() {
		time.Sleep(5 * time.Second)
		log.Println("Connecting to", c.serverURL)
		ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:80/stream?token="+c.clientToken, nil)
		if err != nil {
			log.Println(err.Error())
		}

		go func() {
			for {
				mt, message, err := ws.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
				log.Printf("recv: %s, type: %d", message, mt)
				type test struct {
					Appid    int
					Date     string
					Extras   []byte
					Id       int
					Message  string
					Title    string
					Priority int
				}
				data := test{}
				json.Unmarshal(message, &data)
				log.Println(data.Title, data.Message)
				log.Println(data)
				var postURL = ""
				type discordhook struct {
					Content string `json:"content"`
				}
				var discordSend = discordhook{Content: data.Message}
				log.Println("One", discordSend)
				byteData, err := json.Marshal(&discordSend)
				if err != nil {
					log.Println(err.Error())
				}
				log.Println("Two", string(byteData))
				resp, err := http.Post(postURL, "application/json", bytes.NewReader(byteData))
				if err != nil {
					log.Println(err.Error())
				} else if resp.StatusCode != http.StatusNoContent {
					log.Println(resp.Status)
				}
			}
		}()
		time.Sleep(5 * time.Second)
		c.msgHandler.SendMessage(plugin.Message{
			Message: "The plugin has been enabled for 5 seconds.",
		})
	}()
	return nil
}

// Disable disables the plugin.
func (c *MyPlugin) Disable() error {
	return nil
}

func (c *MyPlugin) GetDisplay(location *url.URL) string {
	var toReturn = ""

	if c.userCtx.Admin {
		toReturn += "Greetings Administrator "
	} else {
		toReturn += "Greatings "
	}
	toReturn += c.userCtx.Name + "\n"

	return toReturn
}

func (c *MyPlugin) SetMessageHandler(h plugin.MessageHandler) {
	c.msgHandler = h
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &MyPlugin{userCtx: ctx, serverURL: "http://localhost/", clientToken: "Ct7UrVHyPQRuwcp"}
}

func main() {
	panic("this should be built as go plugin")
}
