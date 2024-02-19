package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gotify/plugin-api"
)

var info = plugin.Info{
	ModulePath:  "github.com/CEKlopfenstein/gotify-repeater",
	Version:     "2024.0.4",
	Author:      "CEKlopfenstein",
	Description: "A simple Plugin that provides the ability to pass notifications recieved throught to discord. (Current Implementation. More planned.)",
	Name:        "Gotify Repeater",
}

// GetGotifyPluginInfo returns gotify plugin info.
func GetGotifyPluginInfo() plugin.Info {
	return info
}

// GotifyRepeaterPlugin is the gotify plugin instance.
type GotifyRepeaterPlugin struct {
	userCtx  plugin.UserContext
	config   *Config
	listener *websocket.Conn
}

type GotifyMessageStruct struct {
	Appid    int
	Date     string
	Extras   []byte
	Id       int
	Message  string
	Title    string
	Priority int
}

type DiscordWebhookPayload struct {
	Content string `json:"content"`
}

func (c *GotifyRepeaterPlugin) StartRepeater() {
	// Kill existing connection if it already is started
	if c.listener != nil {
		listener := c.listener
		c.listener = nil
		listener.Close()
	}

	var attemptTick = -1
	var attemptLimit = 100
	log.Println("Repeater Attempting to Connect")
	for {
		if attemptTick >= attemptLimit {
			log.Println("Repeater Failed to Connect to Server due to refused connection (Slow Start Up Likely)")
			return
		}
		time.Sleep(100 * time.Millisecond)
		attemptTick++
		health, err := url.Parse(c.config.ServerURL)
		if err != nil {
			log.Println(err)
			return
		}
		health.Path = "/version"
		resp, err := http.Get(health.String())
		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Println(resp.Status)
			continue
		}
		break
	}

	u, err := url.Parse(c.config.ServerURL)
	if err != nil {
		log.Println(err)
		return
	}
	u.Path = "/stream"
	query := u.Query()
	query.Add("token", c.config.ClientToken)
	u.RawQuery = query.Encode()
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		log.Println("Invalid Scheme for URL")
		return
	}

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	c.listener = ws
	log.Println("Repeater Connected")
	go func() {
		for {
			gotifyMessage := GotifyMessageStruct{}
			err := ws.ReadJSON(&gotifyMessage)
			if err != nil && c.listener != nil {
				log.Println("Failed to Read in Gotify Message from Stream:", err)
				return
			} else if c.listener == nil {
				return
			}

			var discordPayload = DiscordWebhookPayload{Content: "# " + gotifyMessage.Title + "\n\n" + gotifyMessage.Message}

			discordBytePayload, err := json.Marshal(&discordPayload)
			if err != nil {
				log.Println("Failed To Build Discord Webhook Payload:", err.Error())
				continue
			}
			resp, err := http.Post(c.config.DiscordWebHook, "application/json", bytes.NewReader(discordBytePayload))
			if err != nil {
				log.Println("Failed to Send Discord Webhook:", err.Error())
				continue
			} else if resp.StatusCode != http.StatusNoContent {
				log.Println("Discord Webhook returned response other than 204. Response:", resp.Status)
			}
		}
	}()
}

func (c *GotifyRepeaterPlugin) StopRepeater() {
	if c.listener != nil {
		listener := c.listener
		c.listener = nil
		listener.Close()
	}
}

// Enable enables the plugin.
func (c *GotifyRepeaterPlugin) Enable() error {
	go c.StartRepeater()
	return nil
}

// Disable disables the plugin.
func (c *GotifyRepeaterPlugin) Disable() error {
	c.StopRepeater()
	return nil
}

func (c *GotifyRepeaterPlugin) GetDisplay(location *url.URL) string {
	var toReturn = ""

	toReturn += "Version: " + info.Version + "\n\nDescription: " + info.Description + "\n\n"

	toReturn += "In order to have this plugin function correctly 3 values are needed within. `discordwebhook`, `clienttoken`, and `serverurl`.\n\n`serverurl` can often be left as the default. Unless you enable HTTPS or wish to have the the plugin listen through some other URL. Note this can allow you to have the plugin listen to a different server entirely. This is not advised. As reconnection after a lost connection is not attempted at this time.\n\n`clienttoken` is the client the plugin will connect as. This can be any client you desire. It would be advisable to create it's own client in the Client Menu.\n\n`discordwebhool` is the webhook the plugin will use to send out messages. Currently the name is not modified by the plugin. So the username will be what ever it was set to at creation."
	return toReturn
}

type Config struct {
	DiscordWebHook string
	ClientToken    string
	ServerURL      string
}

// Set Default Values of Config
func (c *GotifyRepeaterPlugin) DefaultConfig() interface{} {
	return &Config{
		DiscordWebHook: "",
		ClientToken:    "",
		ServerURL:      "http://localhost",
	}
}

func (c *GotifyRepeaterPlugin) ValidateAndSetConfig(cd interface{}) error {
	config := cd.(*Config)
	// Validation of Discord Webhook
	if len(config.DiscordWebHook) == 0 {
		return errors.New("discord Webhook required")
	} else {
		resp, err := http.Get(config.DiscordWebHook)
		if err != nil {
			return errors.Join(errors.New("discord Webhook invalid"), err)
		} else if resp.StatusCode != http.StatusOK {
			return errors.New("discord Webhook invalid. Discord returned value other than success")
		}
	}

	// Validation of local server URL
	if len(config.ServerURL) == 0 {
		return errors.New("server url invalid")
	} else {
		u, err := url.Parse(config.ServerURL)
		if err != nil {
			return errors.Join(errors.New("server url invalid"), err)
		}
		switch u.Scheme {
		case "http":
		case "https":
		default:
			return errors.New("server URL invalid URL must be HTTP or HTTPS")
		}
		if len(u.Path) > 0 {
			return errors.New("server URL invalid URL must not include a path")
		}
	}

	if len(config.ClientToken) == 0 {
		return errors.New("client token required")
	}
	c.config = config
	return nil
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &GotifyRepeaterPlugin{userCtx: ctx}
}

func main() {
	panic("this should be built as go plugin")
}
