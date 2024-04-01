package main

import (
	_ "embed"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/CEKlopfenstein/gotify-repeater/config"
	"github.com/CEKlopfenstein/gotify-repeater/relay"
	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/CEKlopfenstein/gotify-repeater/transmitter"
	"github.com/CEKlopfenstein/gotify-repeater/user"
	"github.com/gin-gonic/gin"
	"github.com/gotify/plugin-api"
)

var info = plugin.Info{
	ModulePath:  "github.com/CEKlopfenstein/gotify-repeater",
	Version:     "2024.1.x",
	Author:      "CEKlopfenstein",
	Description: "A simple plugin that acts as a relay to discord. (Current Implementation. More planned.)",
	Name:        "Gotify Relay",
}

// GetGotifyPluginInfo returns gotify plugin info.
func GetGotifyPluginInfo() plugin.Info {
	return info
}

// GotifyRelayPlugin is the gotify plugin instance.
type GotifyRelayPlugin struct {
	userCtx        plugin.UserContext
	config         *structs.Config
	relay          relay.Relay
	basePath       string
	hostName       string
	storageHandler plugin.StorageHandler
}

// Enable enables the plugin.
func (c *GotifyRelayPlugin) Enable() error {
	var server = server.SetupServer(c.hostName, c.config.ClientToken)
	var discord = transmitter.BuildDiscordTransmitter(c.config.DiscordWebHook, info.Name)
	c.relay.SetServer(server)
	c.relay.ClearTransmitFunctions()
	c.relay.AddTransmitter(transmitter.LogTransmittor{})
	c.relay.AddTransmitter(discord)
	go c.relay.Start()
	return nil
}

// Disable disables the plugin.
func (c *GotifyRelayPlugin) Disable() error {
	c.relay.Stop()
	return nil
}

//go:embed SetupHints.md
var setupHints string

func (c *GotifyRelayPlugin) GetDisplay(location *url.URL) string {
	var toReturn = ""

	toReturn += "## Version: " + info.Version + "\n\n## Description:\n" + info.Description + "\n\n"

	toReturn += setupHints

	toReturn += "\n\n## [Config Page](" + c.basePath + ")"

	return toReturn
}

// Set Default Values of Config
func (c *GotifyRelayPlugin) DefaultConfig() interface{} {
	return &structs.Config{
		DiscordWebHook: "",
		ClientToken:    "",
		ServerURL:      "http://localhost",
	}
}

func (c *GotifyRelayPlugin) ValidateAndSetConfig(cd interface{}) error {
	config := cd.(*structs.Config)
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

func (c *GotifyRelayPlugin) RegisterWebhook(basePath string, mux *gin.RouterGroup) {
	c.basePath = basePath
	user.BuildInterface(basePath, mux, &c.relay, c.config, c.storageHandler)
}

func (c *GotifyRelayPlugin) SetStorageHandler(h plugin.StorageHandler) {
	c.storageHandler = h
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	conf := config.Get()

	var host string
	if *conf.Server.SSL.Enabled {
		host = "https://"
	} else {
		host = "http://"
	}
	if *conf.Server.SSL.Enabled && len(conf.Server.SSL.ListenAddr) == 0 {
		host += "127.0.0.1"
	} else if !*conf.Server.SSL.Enabled && len(conf.Server.ListenAddr) == 0 {
		host += "127.0.0.1"
	} else {
		host += conf.Server.ListenAddr
	}
	if *conf.Server.SSL.Enabled && conf.Server.SSL.Port != 443 {
		host += ":" + strconv.Itoa(conf.Server.SSL.Port)
	} else if conf.Server.Port != 80 {
		host += ":" + strconv.Itoa(conf.Server.Port)
	}

	return &GotifyRelayPlugin{userCtx: ctx, hostName: host}
}

func main() {
	panic("this should be built as go plugin")
}
