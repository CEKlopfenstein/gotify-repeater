package relay

import (
	"log"
	"time"

	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/CEKlopfenstein/gotify-repeater/storage"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	transmitter "github.com/CEKlopfenstein/gotify-repeater/transmitters"
	"github.com/gorilla/websocket"
)

type Relay struct {
	listener        *websocket.Conn
	server          server.Server
	senderFunctions map[int]func(structs.GotifyMessageStruct)
	nextFunctionID  int
	transmitters    map[int]transmitter.Transmitter
	storage         storage.Storage
}

func (relay *Relay) SetServer(server server.Server) {
	relay.server = server
}

func (relay *Relay) SetStorage(storage storage.Storage) {
	relay.storage = storage
	relay.loadTransmitters()
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

func (relay *Relay) loadTransmitters() {
	relay.transmitters = map[int]transmitter.Transmitter{}
	var transFromStore = relay.storage.GetTransmitters()
	for key := range transFromStore {
		relay.transmitters[key] = transmitter.RehydrateTransmitter(transFromStore[key])
	}
}

func (relay *Relay) ReloadTransmitters() {
	relay.loadTransmitters()
}

func (relay *Relay) saveTransmitters() {
	transmitters := relay.GetTransmitters()
	var transToStore = map[int]structs.TransmitterStorage{}
	for key := range transmitters {
		transToStore[key] = transmitters[key].GetStorageValue(key)
	}
	relay.storage.SaveTransmitters(transToStore)
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
				if relay.transmitters[key].Active() {
					relay.transmitters[key].Transmit(gotifyMessage, relay.server)
				}
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

func (relay *Relay) AddTransmitter(sender transmitter.Transmitter) int {
	var id = relay.storage.GetCurrentTransmitterNextID()
	relay.storage.AddTransmitter(sender.GetStorageValue(id))
	relay.loadTransmitters()
	return id
}

func (relay *Relay) ClearTransmitters() int {
	if relay.transmitters == nil {
		relay.transmitters = make(map[int]transmitter.Transmitter)
	}
	count := len(relay.transmitters)

	for key := range relay.transmitters {
		delete(relay.transmitters, key)
	}
	relay.saveTransmitters()
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
		relay.transmitters = make(map[int]transmitter.Transmitter)
	}
	delete(relay.transmitters, index)
	relay.saveTransmitters()
}

func (relay *Relay) RemoveTransmitFunction(index int) {
	if relay.senderFunctions == nil {
		relay.senderFunctions = make(map[int]func(structs.GotifyMessageStruct))
	}
	delete(relay.senderFunctions, index)
}

func (relay *Relay) GetTransmitters() map[int]transmitter.Transmitter {
	if relay.transmitters == nil {
		relay.transmitters = make(map[int]transmitter.Transmitter)
	}

	return relay.transmitters
}

func (relay *Relay) SetTransmitterStatus(id int, status bool) {
	relay.transmitters[id].SetStatus(status)
	relay.saveTransmitters()
}

func (relay *Relay) Stop() error {
	relay.saveTransmitters()
	var con = relay.listener
	relay.listener = nil
	var err = con.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error while Closing Connection:", err)
	}
	return err
}
