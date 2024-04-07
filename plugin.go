package main

import (
	_ "embed"
	"log"
	"net/url"
	"strconv"

	"github.com/CEKlopfenstein/gotify-repeater/config"
	"github.com/CEKlopfenstein/gotify-repeater/relay"
	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/CEKlopfenstein/gotify-repeater/storage"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
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
	userCtx  plugin.UserContext
	config   *structs.Config
	relay    relay.Relay
	basePath string
	hostName string
	storage  storage.Storage
}

// Enable enables the plugin.
func (c *GotifyRelayPlugin) Enable() error {
	log.Println("Client Token: ", c.storage.GetClientToken())
	var server = server.SetupServer(c.hostName, c.storage.GetClientToken())
	// var discord = discordTransmitter.BuildDiscordTransmitter("", info.Name, true)
	c.relay.SetServer(server)
	c.relay.SetStorage(c.storage)
	c.relay.ClearTransmitFunctions()
	// c.relay.AddTransmitter(logTransmitter.LogTransmittor{})
	// c.relay.AddTransmitter(&discord)
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

func (c *GotifyRelayPlugin) RegisterWebhook(basePath string, mux *gin.RouterGroup) {
	c.basePath = basePath
	user.BuildInterface(basePath, mux, &c.relay, c.config, c.storage, c.hostName)
}

func (c *GotifyRelayPlugin) SetStorageHandler(h plugin.StorageHandler) {
	c.storage.StorageHandler = h
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
