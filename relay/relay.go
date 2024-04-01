package relay

import (
	"log"
	"time"

	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/gorilla/websocket"
)

type Relay struct {
	listener          *websocket.Conn
	server            server.Server
	senderFunctions   map[int]func(structs.GotifyMessageStruct)
	nextFunctionID    int
	transmitters      map[int]structs.Transmitter
	nextTransmitterID int
}

func (relay *Relay) SetServer(server server.Server) {
	relay.server = server
}

func (relay *Relay) GetServer() server.Server {
	return relay.server
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
			var gotifyMessage = structs.GotifyMessageStruct{}
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
			for key := range relay.transmitters {
				relay.transmitters[key].BuildTransmitterFunction()(gotifyMessage, relay.server)
			}
		}
	}()
}

func (relay *Relay) AddTransmitFunction(sender func(structs.GotifyMessageStruct)) int {
	if relay.senderFunctions == nil {
		relay.senderFunctions = make(map[int]func(structs.GotifyMessageStruct))
	}
	id := relay.nextFunctionID
	relay.nextFunctionID++
	relay.senderFunctions[id] = sender
	return id
}

func (relay *Relay) AddTransmitter(sender structs.Transmitter) int {
	if relay.transmitters == nil {
		relay.transmitters = make(map[int]structs.Transmitter)
	}
	id := relay.nextTransmitterID
	relay.nextTransmitterID++
	relay.transmitters[id] = sender
	return id
}

func (relay *Relay) ClearTransmitters() int {
	if relay.transmitters == nil {
		relay.transmitters = make(map[int]structs.Transmitter)
	}
	count := len(relay.transmitters)

	for key := range relay.transmitters {
		delete(relay.transmitters, key)
	}

	return count
}

func (relay *Relay) ClearTransmitFunctions() int {
	if relay.senderFunctions == nil {
		relay.senderFunctions = make(map[int]func(structs.GotifyMessageStruct))
	}
	count := len(relay.senderFunctions)

	for key := range relay.senderFunctions {
		delete(relay.senderFunctions, key)
	}

	return count
}

func (relay *Relay) RemoveTransmitter(index int) {
	if relay.transmitters == nil {
		relay.transmitters = make(map[int]structs.Transmitter)
	}
	delete(relay.transmitters, index)
}

func (relay *Relay) RemoveTransmitFunction(index int) {
	if relay.senderFunctions == nil {
		relay.senderFunctions = make(map[int]func(structs.GotifyMessageStruct))
	}
	delete(relay.senderFunctions, index)
}

func (relay *Relay) GetTransmitters() map[int]structs.Transmitter {
	if relay.transmitters == nil {
		relay.transmitters = make(map[int]structs.Transmitter)
	}

	return relay.transmitters
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
