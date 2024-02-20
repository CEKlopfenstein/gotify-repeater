# CEKlopfenstein/gotify-repeater

A plugin for [gotify/server](https://github.com/gotify/server) using [gotify/plugin-template](https://github.com/gotify/plugin-template) as a base.

The goal of this plugin is to enable Gotify to send out notifications through services beyond just it's own clients. Initially this will be Discord Webhooks with farther Expanstion in the future.

## But why?
I have a few things that support Gotify for notifications but not something else I use. That and I'd like a centralized places to direct all the notifications within my homelab. Is this the best method? Who knows. But let's have fun doing it.

## Plans (In order of priority)
- [x] Basic Discord Webhook integration
- [ ] User Friendly configuration
- [ ] Embedded Discord Webhook integration
- [ ] Automated Releases
- [ ] More integrations
   - Telegram
   - Pushbullet
   - Discord bot? (likely more complex than needed bug hey. I'm spit balling here.)

## Building
For now please refer to [OG_README.md](OG_README.md) for documentation on how to build.

Please note that modifications have been made to the make file to allow it to function on my machine spesifically. They may not function else where. I plan to eventually remedy that.

## Trial Running
The Makefile has had an extra command added to it to allow for the automated running of a [gotify/server](https://github.com/gotify/server) with the plugin install. To access run `make run`. Note as does the origial Makefile it also requires the installation of docker. It will also use `sudo` to attempt to run docker. You have been warned.