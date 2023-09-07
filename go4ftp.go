package go4ftp

import (
	"errors"
	"fmt"
	"time"
)

type ConnConfig struct {
	Protocol      string
	Host          string
	Port          int
	User          string
	Password      string
	Timeout       time.Duration
	IgnoreHostKey bool
}

type FileUpload struct {
	LocalFilepath string
	FTPFolder     string
	FTPFileName   string
}

type Entries struct {
	Name string `json:"name"`
	Size uint64 `json:"size"`
}

type Instance interface {
	Ping() error
	Connect() error
	Close() error

	Read(string) ([]Entries, error)
	UploadFile(FileUpload) error
}

func NewInstance(config ConnConfig) (Instance, error) {
	if config.Protocol == "sftp" {
		return newSFTP(config), nil
	}
	if config.Protocol == "ftp" {
		return newFTP(config), nil
	}
	return nil, errors.New(fmt.Sprintf("Protocol %s not supported", config.Protocol))
}
