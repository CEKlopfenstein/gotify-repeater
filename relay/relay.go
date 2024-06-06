package relay

import (
	"log"
	"time"

	"github.com/CEKlopfenstein/gotify-repeater/gotify_api"
	"github.com/CEKlopfenstein/gotify-repeater/storage"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/CEKlopfenstein/gotify-repeater/transmitters"
	"github.com/gorilla/websocket"
)

type Relay struct {
	listener     *websocket.Conn
	gotifyApi    gotify_api.GotifyApi
	transmitters map[int]transmitters.Transmitter
	storage      storage.Storage
	userName     string
	logger       *log.Logger
}

func (relay *Relay) SetGotifyApi(gotifyApi gotify_api.GotifyApi) {
	relay.gotifyApi = gotifyApi
}

func (relay *Relay) SetUserName(userName string) {
	relay.userName = userName
}

func (relay *Relay) SetStorage(storage storage.Storage) {
	relay.storage = storage
	relay.loadTransmitters()
}

func (relay *Relay) SetLogger(logger *log.Logger) {
	relay.logger = logger
}
func (relay *Relay) GetGotifyApi() gotify_api.GotifyApi {
	return relay.gotifyApi
}

func (relay *Relay) Start() {
	var attemptTick = -1
	var attemptLimit = 100
	for {
		if attemptTick >= attemptLimit {
			relay.logger.Printf("%s Limit Exceeded for checking HTTP(s) status\n", relay.userName)
			return
		}
		time.Sleep(100 * time.Millisecond)
		attemptTick++
		_, check := relay.gotifyApi.GetServerInfo()
		if check == nil {
			break
		}
		relay.logger.Printf("%s Checking HTTP(s) Status. Error: %s\n", relay.userName, check.Error())
	}
	err := relay.connectToStream()
	if err != nil {
		relay.logger.Printf("%s Failed to make connection to stream. Error: %s\n", relay.userName, err.Error())
		return
	}
	relay.logger.Printf("Relay for %s connected to stream.\n", relay.userName)
	relay.startSender()
}

func (relay *Relay) UpdateToken(token string) error {
	relay.storage.SaveClientToken(token)
	err := relay.gotifyApi.UpdateToken(token)
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
		relay.logger.Printf("%s Active Connection Found. Now closing.\n", relay.userName)
		var err = relay.Stop()
		if err != nil {
			return err
		}
	}
	listener, err := relay.gotifyApi.GetStream()
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
				relay.logger.Printf("Connection Closed for %s Relay.\n", relay.userName)
				return
			}
			if err != nil {
				relay.logger.Printf("Failed to read in Gotify Message from %s Stream: %s\n", relay.userName, err.Error())
				relay.Stop()
				go func() {
					time.Sleep(time.Second * 1)
					go relay.Start()
				}()
			}

			if len(gotifyMessage.Message)+len(gotifyMessage.Title) == 0 {
				continue
			}

			var activeFlag = false

			for key := range relay.transmitters {
				if relay.transmitters[key].Active() {
					activeFlag = true
					relay.transmitters[key].Transmit(gotifyMessage, relay.gotifyApi)
				}
			}

			if activeFlag {
				relay.saveTransmitters()
			}
		}
	}()
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

func (relay *Relay) RemoveTransmitter(index int) {
	if relay.transmitters == nil {
		relay.transmitters = make(map[int]transmitters.Transmitter)
	}
	delete(relay.transmitters, index)
	relay.saveTransmitters()
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
			relay.logger.Printf("Error while closing %s connection: %s\n", relay.userName, err.Error())
		}
		return err
	}
	return nil
}
