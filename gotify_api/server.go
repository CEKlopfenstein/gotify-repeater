package gotify_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type GotifyApi struct {
	serverUrl    string
	client_token string
	logger       *log.Logger
}

type GotifyServerInfo struct {
	BuildDate string
	Commit    string
	Version   string
}

type GotifyApplication struct {
	DefaultPriority int
	Description     string
	Id              int
	Image           string
	Internal        bool
	LastUsed        string
	Name            string
	Token           string
}

func SetupGotifyApi(serverUrl string, token string) GotifyApi {
	return GotifyApi{serverUrl: serverUrl, client_token: token, logger: log.Default()}
}

func SetupGotifyApiExternalLog(serverUrl string, token string, logger *log.Logger) GotifyApi {
	return GotifyApi{serverUrl: serverUrl, client_token: token, logger: logger}
}

func (server *GotifyApi) request(path string, method string, reqBody []byte) ([]byte, error) {
	var body []byte
	versionURL, err := url.Parse(server.serverUrl)
	if err != nil {
		return body, err
	}
	versionURL.Path = path

	var reader io.Reader = nil
	if reqBody != nil {
		reader = bytes.NewReader(reqBody)
	}

	client := http.Client{}
	req, err := http.NewRequest(method, versionURL.String(), reader)
	if err != nil {
		return body, err
	}
	req.Header.Set("X-Gotify-Key", server.client_token)
	if reader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := client.Do(req)
	if err != nil {
		return body, err
	}

	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return body, nil
	}

	if res.StatusCode != http.StatusOK {
		return body, errors.New(res.Status)
	}

	return body, nil
}

func (server *GotifyApi) GetServerInfo() (GotifyServerInfo, error) {
	versionURL, err := url.Parse(server.serverUrl)
	var serverInfo = GotifyServerInfo{}
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

func (server *GotifyApi) GetStream() (*websocket.Conn, error) {
	var streamUrl, err = url.Parse(server.serverUrl)
	if err != nil {
		return nil, err
	}

	streamUrl.Path = "/stream"
	var tokenQuery = streamUrl.Query()
	tokenQuery.Add("token", server.client_token)
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

func (server *GotifyApi) GetApplications() ([]GotifyApplication, error) {
	applications := []GotifyApplication{}
	body, err := server.request("/application", http.MethodGet, nil)
	if err != nil {
		return applications, err
	}
	err = json.Unmarshal(body, &applications)
	if err != nil {
		return applications, err
	}

	return applications, nil
}

func (server *GotifyApi) GetApplication(appId int) (GotifyApplication, error) {
	application := GotifyApplication{}

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

func (server *GotifyApi) CheckToken(token string) error {
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

	return nil
}

func (server *GotifyApi) UpdateToken(token string) error {
	err := server.CheckToken(token)
	if err != nil {
		return err
	}
	server.client_token = token
	return nil
}

type GotifyClientInfo struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

func (server *GotifyApi) FindClientFromToken(token string) GotifyClientInfo {
	body, err := server.request("/client", http.MethodGet, nil)
	if err != nil {
		server.logger.Println(err)
		return GotifyClientInfo{}
	}
	var clients []GotifyClientInfo
	err = json.Unmarshal(body, &clients)
	if err != nil {
		server.logger.Println(err)
		return GotifyClientInfo{}
	}

	for i := 0; i < len(clients); i++ {
		if clients[i].Token == token {
			return clients[i]
		}
	}

	return GotifyClientInfo{}
}

func (server *GotifyApi) DeleteClient(id int) {
	_, err := server.request(fmt.Sprintf("/client/%d", id), http.MethodDelete, nil)
	if err != nil {
		server.logger.Println(err)
		return
	}
}

func (server *GotifyApi) CreateClient(name string) (GotifyClientInfo, error) {
	type newClient struct {
		Name string `json:"name"`
	}
	reqBody, err := json.Marshal(newClient{Name: name})

	if err != nil {
		return GotifyClientInfo{}, err
	}

	body, err := server.request("/client", http.MethodPost, reqBody)
	if err != nil {
		return GotifyClientInfo{}, err
	}

	var client GotifyClientInfo
	err = json.Unmarshal(body, &client)
	if err != nil {
		return client, err
	}

	return client, nil
}
