package bmcftp

import (
	"errors"
	"fmt"
	"strings"
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
	UploadFile(FileUpload) error
}

func NewInstance(config ConnConfig) (Instance, error) {
	fmt.Printf("Creating new %s Instance\n", strings.ToUpper(config.Protocol))
	if config.Protocol == "sftp" {
		return NewSFTP(config), nil
	}
	if config.Protocol == "ftp" {
		return NewFTP(config), nil
	}
	return nil, errors.New(fmt.Sprintf("Protocol %s not supported", config.Protocol))
}
