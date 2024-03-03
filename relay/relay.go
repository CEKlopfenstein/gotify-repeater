package relay

import (
	"log"
	"time"

	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/gorilla/websocket"
)

type Relay struct {
	listener      *websocket.Conn
	server        server.Server
	sendFunctions []func(GotifyMessageStruct)
}

type GotifyMessageStruct struct {
	Appid    int
	Date     string
	Id       int
	Message  string
	Title    string
	Priority int
}

func (repeater *Relay) SetServer(server server.Server) {
	repeater.server = server
}

func (repeater *Relay) Start() {
	var attemptTick = -1
	var attemptLimit = 100
	for {
		if attemptTick >= attemptLimit {
			log.Println("Limit Exceeded for checking HTTP(s) status")
			return
		}
		time.Sleep(100 * time.Millisecond)
		attemptTick++
		_, check := repeater.server.GetServerInfo()
		if check == nil {
			break
		}
		log.Println("Checking HTTP(s) Status. Error:", check)
	}
	err := repeater.connectToStream()
	if err != nil {
		log.Println("Failed to make connection to stream. Error:", err)
		return
	}
	log.Println("Connected to stream")
	repeater.startSender()
}

func (repeater *Relay) connectToStream() error {
	if repeater.listener != nil {
		log.Println("Active Connection Found Closing")
		var err = repeater.Stop()
		if err != nil {
			return err
		}
	}
	listener, err := repeater.server.GetStream()
	if err != nil {
		return err
	}
	repeater.listener = listener
	return nil
}

func (repeater *Relay) startSender() {
	go func() {
		var con = repeater.listener
		defer con.Close()
		for {
			var gotifyMessage = GotifyMessageStruct{}
			var err = con.ReadJSON(&gotifyMessage)
			if repeater.listener == nil {
				log.Println("Connection Closed")
				return
			}
			if err != nil {
				log.Println("Failed to Read in Gotify Message from Stream:", err)
			}
			for sender := 0; sender < len(repeater.sendFunctions); sender++ {
				repeater.sendFunctions[sender](gotifyMessage)
			}
		}
	}()
}

func (repeater *Relay) AddSender(sender func(GotifyMessageStruct)) int {
	repeater.sendFunctions = append(repeater.sendFunctions, sender)
	return len(repeater.sendFunctions) - 1
}

func (repeater *Relay) ClearSenders() int {
	var count = len(repeater.sendFunctions)

	repeater.sendFunctions = []func(GotifyMessageStruct){}

	return count
}

func (repeater *Relay) RemoveSender(index int) {
	var newSendersArray = []func(GotifyMessageStruct){}
	for senderIndex := 0; senderIndex < len(repeater.sendFunctions); senderIndex++ {
		if senderIndex != index {
			newSendersArray = append(newSendersArray, repeater.sendFunctions[senderIndex])
		} else {
			newSendersArray = append(newSendersArray, func(msg GotifyMessageStruct) {
				// Blank to preserve indexes already returned.
			})
		}
	}
	repeater.sendFunctions = newSendersArray
}

func (repeater *Relay) Stop() error {
	var con = repeater.listener
	repeater.listener = nil
	var err = con.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error while Closing Connection:", err)
	}
	return err
}
