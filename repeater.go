package main

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Repeater struct {
	listener      *websocket.Conn
	host          string
	streamUrl     string
	token         string
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

func (repeater *Repeater) SetUrlAndToken(url string, token string) error {
	repeater.host = url
	repeater.token = token
	return repeater.buildStreamUrl()
}

func (repeater *Repeater) Start() {
	var attemptTick = -1
	var attemptLimit = 100
	for {
		if attemptTick >= attemptLimit {
			log.Println("Limit Exceeded for checking HTTP(s) status")
			return
		}
		time.Sleep(100 * time.Millisecond)
		attemptTick++
		var check = repeater.checkForAcceptingConnections()
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

func (repeater *Repeater) connectToStream() error {
	if repeater.listener != nil {
		log.Println("Active Connection Found Closing")
		var err = repeater.Stop()
		if err != nil {
			return err
		}
	}
	listener, _, err := websocket.DefaultDialer.Dial(repeater.streamUrl, nil)
	if err != nil {
		return err
	}
	repeater.listener = listener
	return nil
}

func (repeater *Repeater) startSender() {
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

func (repeater *Repeater) AddSender(sender func(GotifyMessageStruct)) int {
	repeater.sendFunctions = append(repeater.sendFunctions, sender)
	return len(repeater.sendFunctions) - 1
}

func (repeater *Repeater) ClearSenders() int {
	var count = len(repeater.sendFunctions)

	repeater.sendFunctions = []func(GotifyMessageStruct){}

	return count
}

func (repeater *Repeater) RemoveSender(index int) {
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

func (repeater *Repeater) Stop() error {
	var con = repeater.listener
	repeater.listener = nil
	var err = con.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error while Closing Connection:", err)
	}
	return err
}

func (repeater *Repeater) checkForAcceptingConnections() error {
	health, err := url.Parse(repeater.host)
	if err != nil {
		return err
	}
	health.Path = "/version"

	resp, err := http.Get(health.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func (repeater *Repeater) buildStreamUrl() error {
	var streamUrl, err = url.Parse(repeater.host)
	if err != nil {
		return err
	}

	streamUrl.Path = "/stream"
	var tokenQuery = streamUrl.Query()
	tokenQuery.Add("token", repeater.token)
	streamUrl.RawQuery = tokenQuery.Encode()

	switch streamUrl.Scheme {
	case "http":
		streamUrl.Scheme = "ws"
	case "https":
		streamUrl.Scheme = "wss"
	default:
		return errors.New("invalid Schema in use in host URL")
	}

	repeater.streamUrl = streamUrl.String()

	return nil
}
