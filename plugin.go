package main

import (
	"bytes"
	_ "embed"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/CEKlopfenstein/gotify-repeater/config"
	"github.com/CEKlopfenstein/gotify-repeater/gotify_api"
	"github.com/CEKlopfenstein/gotify-repeater/relay"
	"github.com/CEKlopfenstein/gotify-repeater/storage"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/CEKlopfenstein/gotify-repeater/transmitters"
	"github.com/CEKlopfenstein/gotify-repeater/user_interface"
	"github.com/gin-gonic/gin"
	"github.com/gotify/plugin-api"
)

var VERSION string

//go:embed MAJOR_CHANGELOG.md
var changesSinceLastFullRelease string

var info = plugin.Info{
	ModulePath:  "github.com/CEKlopfenstein/gotify-repeater",
	Version:     "BLANK",
	Author:      "CEKlopfenstein",
	Description: "A simple plugin that acts as a relay to discord. (The fun is just beginning more to come.)",
	Name:        "Gotify Relay",
}

// GetGotifyPluginInfo returns gotify plugin info.
func GetGotifyPluginInfo() plugin.Info {
	if len(VERSION) > 0 {
		info.Version = VERSION
	}
	return info
}

// GotifyRelayPlugin is the gotify plugin instance.
type GotifyRelayPlugin struct {
	userCtx   plugin.UserContext
	config    *structs.Config
	relay     relay.Relay
	basePath  string
	hostName  string
	storage   storage.Storage
	enabled   bool
	logger    *log.Logger
	logBuffer *bytes.Buffer
}

// Enable enables the plugin.
func (c *GotifyRelayPlugin) Enable() error {
	c.enabled = true
	var server = gotify_api.SetupGotifyApiExternalLog(c.hostName, c.storage.GetClientToken(), c.logger)
	c.relay.SetUserName(c.userCtx.Name)
	c.relay.SetGotifyApi(server)
	c.relay.SetLogger(c.logger)
	c.relay.SetStorage(c.storage)
	go c.relay.Start()
	c.logger.Printf("Plugin Enabled for %s\n", c.userCtx.Name)
	return nil
}

// Disable disables the plugin.
func (c *GotifyRelayPlugin) Disable() error {
	c.enabled = false
	c.relay.Stop()
	c.logger.Printf("Plugin Disabled for %s\n", c.userCtx.Name)
	return nil
}

func (c *GotifyRelayPlugin) GetDisplay(location *url.URL) string {
	var toReturn = ""

	toReturn += "## Version: " + info.Version + "\n\n## Description:\n" + info.Description + "\n\n"

	if len(c.storage.GetClientToken()) == 0 {
		toReturn += "Missing Token. Go to Config Page to setup.\n\n"
	}

	toReturn += "## [Config Page](" + c.basePath + ")"
	if !c.enabled {
		toReturn += " is only accessible if plugin is enabled.\n\n"
	}

	toReturn += "\n\n## Change Log\n\nSince last full release.\n\n" + changesSinceLastFullRelease

	return toReturn
}

func (c *GotifyRelayPlugin) RegisterWebhook(basePath string, mux *gin.RouterGroup) {
	c.basePath = basePath
	user_interface.BuildInterface(basePath, mux, &c.relay, c.config, c.storage, c.hostName, c.logger, c.logBuffer)
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

	logBuffer := &(bytes.Buffer{})
	logger := log.New(io.MultiWriter(os.Stdout, logBuffer), "Gotify Relay: ", log.LstdFlags|log.Lmsgprefix)
	logger.Printf("Logger Successfully Created for %s", ctx.Name)

	toReturn := &GotifyRelayPlugin{userCtx: ctx, hostName: host, logger: logger, logBuffer: logBuffer}
	toReturn.storage.Logger = logger
	for tType := range transmitters.Types {
		transmitters.Types[tType].SetGlobalLogger(logger)
	}

	return toReturn
}

func main() {
	panic("this should be built as go plugin")
}
