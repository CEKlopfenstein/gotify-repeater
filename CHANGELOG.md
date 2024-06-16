## 2024.2.28
### Changes:
- Modified method of setting version number. Now provided at compile time through ldflags
- Built and tested basic Release Generation Pipelines for both Prereleases and Proper releases.

## 2024.1.22
### Changes:
- Corrected Various Bugs
    - Includes a bug when handling notifications from Proxmox
- Added ability for the plugin to attempt to auto-recover
- Changed the method of configuration from the use of Configurator to a custom interface.
- Added ability to create multiple "transmitters" on a single relay.
- Added ability to deactivate and activate "transmitters"
- Added ability to delete "transmitters"
- Added "Log Transmitter".
    - Currently only useful for possible debugging and as a demonstration that multiple different types of "transmitters" are possible.

## 2024.0.6
An Initial Base Functionality Use/Proof of Concept. Fully functional in the simplest sense.

* Implemented Discord Webhook
  * The username is "Gotify Repeater" or the name of the Application receiving the notification.
  * Message is Gotify Message title as a "Header" with the proper message below. (Makes use of Discord Messages' new Header Support)
* Configurator Implemented
  * ClientToken - Client that the plugin will listen to. (May belong to another user if desired.)
  * ServerURL - Defaults to `http://localhost`. If you are deploying Gotify outside of docker or have HTTPS configured this may need to be modified. Should be the URL that accesses Gotify on the local machine it is running on.
  * DiscordWebhook - The Discord Webhook that the messages are sent to.
* Displayer contains information on the Configurator values.

## 2024.2.36
Improvements in automated build pipeline. Refactoring and Bugfixes. With a sprinkling of new features and functionality.

### General Improvements
- Modified method of setting version number. Now provided at compile time through ldflags
- Built and tested basic Release Generation Pipelines for both Prereleases and Proper releases.
- Discord Webhooks how hidden unless hovered over.
- Refactored internal package names for clarity.
- Version of the change log is now present within the Plugin Info Page.
    - Limited to changes since last non prerelease. (Contents of MAJOR_CHANGELOG.md in repo)
- Updated README.md
  - Removed certain planned features.

### Bug Fixes
- Added block to prevent spamming out transmittions that include blank titles and messages.
- Corrected reconnect protocal to function as expected.
    - Broken connections will not be reestablished until 1 second after connection was lost.

### New Features
- Added ability to have transmit counts.
    - Updated within the GUI once every 5 seconds currently.
- Logs for the plugin spesifically are now visible within the UI.
    - Refreshes every 5 seconds
- Implemented New Relay Transmitters.
    - Discord Advance
        - Makes use of Embed elements in Discord messages.
        - Currently provides limited benefit over normal Discord Transmitter.
    - Pushbullet