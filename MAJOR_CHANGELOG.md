### Changes:
- Modified method of setting version number. Now provided at compile time through ldflags
- Built and tested basic Release Generation Pipelines for both Prereleases and Proper releases.
- Added block to prevent spamming out transmittions that include blank titles and messages.
- Corrected reconnect protocal to function as expected.
    - Broken connections will not be reestablished until 1 second after connection was lost.