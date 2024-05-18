# CEKlopfenstein/gotify-repeater "Gotify Relay"

[![Release Pipeline](https://github.com/CEKlopfenstein/gotify-repeater/actions/workflows/release.yml/badge.svg?branch=master)](https://github.com/CEKlopfenstein/gotify-repeater/actions/workflows/release.yml)

A "simple" [Gotify](https://gotify.net/) plugin that allows for the relaying of notifications from a Gotify Server to other services/endpoints.

## Features
- Graphical User Interface
   - Manage relay "transmitters"
- Supports Discord Webhooks as a "transmitter" endpoint.

## Motivation
I have a few things that support Gotify for notifications but not something else I use. That and I'd like a centralized places to direct all the notifications within my homelab. Is this the best method? Who knows. But let's have fun doing it.

## Currently Planned Features
- Advance Discord Webhook integration
- More "Transmitter" Options
   - Secondary Gotify Instance
   - Telegram
   - Pushbullet
   - Discord bot? (Extremely unlikely)
- Forwarding Filters

## [Changelog](/CHANGELOG.md)

## Installation
1. Download the [latest "stable" version](https://github.com/CEKlopfenstein/gotify-repeater/releases/latest) for your desired deployment.
   > Note: As of writing I only actively test the AMD64 build. ARM64 and ARM-7 should work. But are not garenteed.
2. Place the downloaded *.so file within your Gotify instance's plugins folder.
   > Default configuration for Docker will find it at `/app/data/plugins/` within the container.
   
   > Limited to Linux and MacOS. [Gotify documentation mentioning the limitation](https://gotify.net/docs/plugin)
3. Start/Restart your Gotify instance. (Required for the plugin to be loaded.)
4. Login to your Gotify instance and navigate to plugs.
   > ![](/images/plugins.png)
5. Enable the Gotify Relay plugin and navigate to the plugin info page.
   > ![](/images/info.png)
6. Click either the Route Prefix or the Config Page link.
7. Click either Use Current Client Token or Create Custom Token.
   > ![](/images/plugin_config.png)
   
   > Use Current Token will have the "relay" attach and listen using the currently logged in token. Create Custom Token will create a new token that the plugin will attempt to manage.
8. Add transmitters as desired using the UI.
   > Transmitters can also be disabled and deleted from this view.

## Building From Source
For now please refer to [OG_README.md](OG_README.md) for documentation on how to build. Cloning the repository and running `make build` "should" work. But is not guarenteed. As it was modified to function on my machine due to some strange issues. And due to `docker` not being configured to be accessible without `sudo` on my machine.

Within the Makefile is also an option to run `make run` which will build the AMD64 version of the plugin and deploy it onto a local instance of Gotify in a Docker container using [gotify/server](https://github.com/gotify/server).

## Other Notes
The original [README](/OG_README.md) from [gotify/plugin-templates](https://github.com/gotify/plugin-template) is avaliable as `OG_README.md`.