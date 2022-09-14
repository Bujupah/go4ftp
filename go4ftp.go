package go4ftp

import (
	"errors"
	"fmt"
	"time"
)

type ConnConfig struct {
	Protocol      string
	Host          string
	Port          string
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

type Instance interface {
	Ping() error
	Connect() error
	Close() error
	UploadFile(FileUpload) error
}

func NewInstance(config ConnConfig) (Instance, error) {
	if config.Protocol == "sftp" {
		return NewSFTP(config), nil
	}
	if config.Protocol == "ftp" {
		return NewFTP(config), nil
	}
	return nil, errors.New(fmt.Sprintf("Protocol %s not supported", config.Protocol))
}
