package storage

import (
	"encoding/json"
	"log"

	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/gotify/plugin-api"
)

type Storage struct {
	StorageHandler plugin.StorageHandler
	innerStore     innerStorageStruct
}

type innerStorageStruct struct {
	Contact      Contact
	ClientToken  string
	Transmitters []structs.TransmitterStorage
}

type Contact struct {
	FirstName string
	LastName  string
	Email     string
}

// Saves the current inner storage struct. Should be called after every save/set
func (storage *Storage) save() {
	storageBytes, _ := json.Marshal(storage.innerStore)
	storage.StorageHandler.Save(storageBytes)
}

// Loads the stored values from the DB into the current inner storage struct. Should be called before every get.
func (storage *Storage) load() {
	storageBytes, err := storage.StorageHandler.Load()
	if err != nil {
		log.Println(err)
		return
	}

	if len(storageBytes) == 0 {
		storageBytes, _ = json.Marshal(storage.innerStore)
		storage.StorageHandler.Save(storageBytes)
	} else {
		json.Unmarshal(storageBytes, &storage.innerStore)
	}
}

func (storage *Storage) GetContact() Contact {
	storage.load()
	return storage.innerStore.Contact
}

func (storage *Storage) SaveContact(contact Contact) {
	storage.innerStore.Contact = contact
	storage.save()
}

func (storage *Storage) GetClientToken() string {
	storage.load()
	return storage.innerStore.ClientToken
}

func (storage *Storage) SaveClientToken(token string) {
	storage.innerStore.ClientToken = token
	storage.save()
}

func (storage *Storage) GetTransmitters() []structs.TransmitterStorage {
	storage.load()
	return storage.innerStore.Transmitters
}

func (storage *Storage) SaveTransmitters(transmitters []structs.TransmitterStorage) {
	storage.innerStore.Transmitters = transmitters
	storage.save()
}
