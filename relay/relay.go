package relay

import (
	"log"
	"time"

	"github.com/CEKlopfenstein/gotify-repeater/server"
	"github.com/CEKlopfenstein/gotify-repeater/storage"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/CEKlopfenstein/gotify-repeater/transmitters"
	"github.com/gorilla/websocket"
)

type Relay struct {
	listener        *websocket.Conn
	server          server.Server
	senderFunctions map[int]func(structs.GotifyMessageStruct)
	nextFunctionID  int
	transmitters    map[int]transmitters.Transmitter
	storage         storage.Storage
	userName        string
}

func (relay *Relay) SetServer(server server.Server) {
	relay.server = server
}

func (relay *Relay) SetUserName(userName string) {
	relay.userName = userName
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
			log.Printf("%s Limit Exceeded for checking HTTP(s) status\n", relay.userName)
			return
		}
		time.Sleep(100 * time.Millisecond)
		attemptTick++
		_, check := relay.server.GetServerInfo()
		if check == nil {
			break
		}
		log.Printf("%s Checking HTTP(s) Status. Error: %s\n", relay.userName, check.Error())
	}
	err := relay.connectToStream()
	if err != nil {
		log.Printf("%s Failed to make connection to stream. Error: %s\n", relay.userName, err.Error())
		return
	}
	log.Printf("Relay for %s connected to stream.\n", relay.userName)
	relay.startSender()
}

func (relay *Relay) UpdateToken(token string) error {
	relay.storage.SaveClientToken(token)
	err := relay.server.UpdateToken(token)
	if err != nil {
		return err
	}
	err = relay.connectToStream()
	if err != nil {
		return err
	}
	return nil
}

func (relay *Relay) connectToStream() error {
	if relay.listener != nil {
		log.Printf("%s Active Connection Found. Now closing.\n", relay.userName)
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
	relay.transmitters = map[int]transmitters.Transmitter{}
	var transFromStore = relay.storage.GetTransmitters()
	for key := range transFromStore {
		relay.transmitters[key] = transmitters.RehydrateTransmitter(transFromStore[key])
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
				log.Printf("Connection Closed for %s Relay.\n", relay.userName)
				return
			}
			if err != nil {
				log.Printf("Failed to read in Gotify Message from %s Stream: %s\n", relay.userName, err.Error())
				relay.Stop()
				go func() {
					time.Sleep(time.Second * 1)
					go relay.Start()
				}()
			}

			if len(gotifyMessage.Message)+len(gotifyMessage.Title) == 0 {
				continue
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

func (relay *Relay) AddTransmitter(sender transmitters.Transmitter) int {
	var id = relay.storage.GetCurrentTransmitterNextID()
	relay.storage.AddTransmitter(sender.GetStorageValue(id))
	relay.loadTransmitters()
	return id
}

func (relay *Relay) ClearTransmitters() int {
	if relay.transmitters == nil {
		relay.transmitters = make(map[int]transmitters.Transmitter)
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
		relay.transmitters = make(map[int]transmitters.Transmitter)
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

func (relay *Relay) GetTransmitters() map[int]transmitters.Transmitter {
	if relay.transmitters == nil {
		relay.transmitters = make(map[int]transmitters.Transmitter)
	}

	return relay.transmitters
}

func (relay *Relay) SetTransmitterStatus(id int, status bool) {
	relay.transmitters[id].SetStatus(status)
	relay.saveTransmitters()
}

func (relay *Relay) Stop() error {
	relay.saveTransmitters()
	if relay.listener != nil {
		var con = relay.listener
		relay.listener = nil
		var err = con.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Error while closing %s connection: %s\n", relay.userName, err.Error())
		}
		return err
	}
	return nil
}
