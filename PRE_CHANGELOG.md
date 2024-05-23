### Changes:
- Added block to prevent spamming out transmittions that include blank titles and messages.
- Corrected reconnect protocal to function as expected.
    - Broken connections will not be reestablished until 1 second after connection was lost.
- Refactored internal package names for clarity.
- Version of the change log is now present within the Plugin Info Page.
    - Limited to changes since last non prerelease. (Contents of MAJOR_CHANGELOG.md in repo)
- Discord Webhooks how hidden unless hovered over.
- Added ability to have transmit counts.
    - Updated within the GUI once every 5 seconds currently.