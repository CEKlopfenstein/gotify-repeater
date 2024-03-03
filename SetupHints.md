
## Setup Hints
In order to have this plugin function correctly 3 values are needed within. `discordwebhook`, `clienttoken`, and `serverurl`.

`serverurl` can often be left as the default. Unless you enable HTTPS or wish to have the the plugin listen through some other URL. Note this can allow you to have the plugin listen to a different server entirely. This is not advised. As reconnection after a lost connection is not attempted at this time.

`clienttoken` is the client the plugin will connect as. This can be any client you desire. It would be advisable to create it's own client in the Client Menu.

`discordwebhook` is the webhook the plugin will use to send out messages. The Webhooks Username will be the name of the application that sent the message. If this fails for any reason it will be the default Username from the Webhook. If that for some reason fails it will be the Plugin Name seen in the Plugin Info.