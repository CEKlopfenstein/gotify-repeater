package main

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/CEKlopfenstein/gotify-repeater/relay"
	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/CEKlopfenstein/gotify-repeater/transmitter"
	"github.com/gotify/plugin-api"
)

var info = plugin.Info{
	ModulePath:  "github.com/CEKlopfenstein/gotify-repeater",
	Version:     "2024.1.x",
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
	userCtx plugin.UserContext
	config  *Config
	relay   relay.Relay
}

// Enable enables the plugin.
func (c *GotifyRepeaterPlugin) Enable() error {
	var server = server.SetupServer(c.config.ServerURL, c.config.ClientToken)
	var discord = transmitter.BuildDiscordTransmitter(server, c.config.DiscordWebHook)
	discord.Username = info.Name
	c.relay.SetServer(server)
	c.relay.ClearSenders()
	c.relay.AddSender(func(msg relay.GotifyMessageStruct) {
		log.Println(msg)
	})
	c.relay.AddSender(discord.BuildTransmitterFunction())
	go c.relay.Start()
	return nil
}

// Disable disables the plugin.
func (c *GotifyRepeaterPlugin) Disable() error {
	c.relay.Stop()
	return nil
}

func (c *GotifyRepeaterPlugin) GetDisplay(location *url.URL) string {
	var toReturn = ""

	toReturn += "Version: " + info.Version + "\n\nDescription: " + info.Description + "\n\n"

	toReturn += "In order to have this plugin function correctly 3 values are needed within. `discordwebhook`, `clienttoken`, and `serverurl`.\n\n`serverurl` can often be left as the default. Unless you enable HTTPS or wish to have the the plugin listen through some other URL. Note this can allow you to have the plugin listen to a different server entirely. This is not advised. As reconnection after a lost connection is not attempted at this time.\n\n`clienttoken` is the client the plugin will connect as. This can be any client you desire. It would be advisable to create it's own client in the Client Menu.\n\n`discordwebhool` is the webhook the plugin will use to send out messages. The Webhooks Username will be the name of the application that. If this fails for any reason it will be the Plugin Name seen in the Plugin Info."

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
