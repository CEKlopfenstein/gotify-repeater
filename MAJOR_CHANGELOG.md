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