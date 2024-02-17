package main

import (
	"net/url"
	"time"

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
	userCtx    plugin.UserContext
	msgHandler plugin.MessageHandler
}

// Enable enables the plugin.
func (c *MyPlugin) Enable() error {
	go func() {
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
	return &MyPlugin{userCtx: ctx}
}

func main() {
	panic("this should be built as go plugin")
}
