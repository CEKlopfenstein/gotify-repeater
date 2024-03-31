package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type Server struct {
	serverUrl string
	token     string
}

type ServerInfo struct {
	BuildDate string
	Commit    string
	Version   string
}

type Application struct {
	DefaultPriority int
	Description     string
	Id              int
	Image           string
	Internal        bool
	LastUsed        string
	Name            string
	Token           string
}

func SetupServer(serverUrl string, token string) Server {
	return Server{serverUrl: serverUrl, token: token}
}

func (server *Server) GetServerInfo() (ServerInfo, error) {
	versionURL, err := url.Parse(server.serverUrl)
	var serverInfo = ServerInfo{}
	if err != nil {
		return serverInfo, err
	}
	versionURL.Path = "/version"
	resp, err := http.Get(versionURL.String())
	if err != nil {
		return serverInfo, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return serverInfo, errors.New(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return serverInfo, err
	}

	err = json.Unmarshal(body, &serverInfo)
	if err != nil {
		return serverInfo, err
	}

	return serverInfo, nil

}

func (server *Server) GetStream() (*websocket.Conn, error) {
	var streamUrl, err = url.Parse(server.serverUrl)
	if err != nil {
		return nil, err
	}

	streamUrl.Path = "/stream"
	var tokenQuery = streamUrl.Query()
	tokenQuery.Add("token", server.token)
	streamUrl.RawQuery = tokenQuery.Encode()

	switch streamUrl.Scheme {
	case "http":
		streamUrl.Scheme = "ws"
	case "https":
		streamUrl.Scheme = "wss"
	default:
		return nil, errors.New("invalid Schema in use in host URL")
	}

	listener, _, err := websocket.DefaultDialer.Dial(streamUrl.String(), nil)
	if err != nil {
		return listener, err
	}
	return listener, nil
}

func (server *Server) GetApplications() ([]Application, error) {
	applications := []Application{}
	applicationURL, err := url.Parse(server.serverUrl)
	if err != nil {
		return applications, err
	}
	applicationURL.Path = "/application"
	query := applicationURL.Query()
	query.Add("token", server.token)
	applicationURL.RawQuery = query.Encode()
	resp, err := http.Get(applicationURL.String())
	if err != nil {
		return applications, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return applications, err
	}
	err = json.Unmarshal(body, &applications)
	if err != nil {
		return applications, err
	}

	return applications, nil
}

func (server *Server) GetApplication(appId int) (Application, error) {
	application := Application{}

	applications, err := server.GetApplications()
	if err != nil {
		return application, err
	}

	for i := 0; i < len(applications); i++ {
		if applications[i].Id == appId {
			return applications[i], nil
		}
	}

	return application, fmt.Errorf("application with id of %d not found", appId)
}

func (server *Server) CheckToken(token string) error {
	currentUserURL, err := url.Parse(server.serverUrl)
	if err != nil {
		return err
	}
	currentUserURL.Path = "/current/user"
	query := currentUserURL.Query()
	query.Add("token", token)
	currentUserURL.RawQuery = query.Encode()
	resp, err := http.Get(currentUserURL.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	log.Println("Tick", string(body))

	return nil
}
