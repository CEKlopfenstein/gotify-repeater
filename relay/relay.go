package relay

import (
	"log"
	"time"

	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/gorilla/websocket"
)

type Relay struct {
	listener        *websocket.Conn
	server          server.Server
	senderFunctions map[int]func(GotifyMessageStruct)
	nextID          int
}

type GotifyMessageStruct struct {
	Appid    int
	Date     string
	Id       int
	Message  string
	Title    string
	Priority int
}

func (relay *Relay) SetServer(server server.Server) {
	relay.server = server
}

func (relay *Relay) Start() {
	var attemptTick = -1
	var attemptLimit = 100
	for {
		if attemptTick >= attemptLimit {
			log.Println("Limit Exceeded for checking HTTP(s) status")
			return
		}
		time.Sleep(100 * time.Millisecond)
		attemptTick++
		_, check := relay.server.GetServerInfo()
		if check == nil {
			break
		}
		log.Println("Checking HTTP(s) Status. Error:", check)
	}
	err := relay.connectToStream()
	if err != nil {
		log.Println("Failed to make connection to stream. Error:", err)
		return
	}
	log.Println("Connected to stream")
	relay.startSender()
}

func (relay *Relay) connectToStream() error {
	if relay.listener != nil {
		log.Println("Active Connection Found Closing")
		var err = relay.Stop()
		if err != nil {
			return err
		}
	}
	listener, err := relay.server.GetStream()
	if err != nil {
		return err
	}
	relay.listener = listener
	return nil
}

func (relay *Relay) startSender() {
	go func() {
		var con = relay.listener
		defer con.Close()
		for {
			var gotifyMessage = GotifyMessageStruct{}
			var err = con.ReadJSON(&gotifyMessage)
			if relay.listener == nil {
				log.Println("Connection Closed")
				return
			}
			if err != nil {
				log.Println("Failed to Read in Gotify Message from Stream:", err)
				relay.Stop()
				relay.Start()
			}

			for key := range relay.senderFunctions {
				relay.senderFunctions[key](gotifyMessage)
			}
		}
	}()
}

func (relay *Relay) AddSender(sender func(GotifyMessageStruct)) int {
	if relay.senderFunctions == nil {
		relay.senderFunctions = make(map[int]func(GotifyMessageStruct))
	}
	id := relay.nextID
	relay.nextID++
	relay.senderFunctions[id] = sender
	return id
}

func (relay *Relay) ClearSenders() int {
	if relay.senderFunctions == nil {
		relay.senderFunctions = make(map[int]func(GotifyMessageStruct))
	}
	count := len(relay.senderFunctions)

	for key := range relay.senderFunctions {
		delete(relay.senderFunctions, key)
	}

	return count
}

func (relay *Relay) RemoveSender(index int) {
	if relay.senderFunctions == nil {
		relay.senderFunctions = make(map[int]func(GotifyMessageStruct))
	}
	delete(relay.senderFunctions, index)
}

func (relay *Relay) Stop() error {
	var con = relay.listener
	relay.listener = nil
	var err = con.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error while Closing Connection:", err)
	}
	return err
}
