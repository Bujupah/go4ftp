package bmcftp

import (
	"errors"
	"fmt"
	"github.com/jlaffaye/ftp"
	"os"
	"path/filepath"
	"strings"
)

type FTP struct {
	config ConnConfig
}

func NewFTP(config ConnConfig) Instance {
	return &FTP{config}
}

func (s *FTP) Ping() error {
	fmt.Printf("Trying to connect to %s server\n", strings.ToUpper(s.config.Type))
	client, err := s.connect()
	if err != nil {
		return err
	}
	defer client.Quit()
	fmt.Printf("Successfully connected to %s server\n", strings.ToUpper(s.config.Type))
	return nil
}

func (s *FTP) UploadFile(fileUpload FileUpload) error {
	client, err := s.connect()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to connect to server: %s", err.Error()))
	}
	defer client.Logout()
	defer client.Quit()

	// Recursively create folder in FTP server
	folders := strings.Split(fileUpload.FTPFolder, "/")
	currentDir, _ := client.CurrentDir()
	for _, folder := range folders {

		current, _ := client.CurrentDir()
		current = filepath.Join(current, folder)

		if err := client.ChangeDir(current); err != nil {
			// create folder
			fmt.Printf("Navigating to %s and creating folder %s\n", current, folder)
			if err := client.MakeDir(folder); err != nil {
				return errors.New(fmt.Sprintf("Failed to create folder %s: %s", folder, err.Error()))
			}
		}
		// change to folder
		if err := client.ChangeDir(current); err != nil {
			return errors.New(fmt.Sprintf("Failed to change to folder %s: %s", folder, err.Error()))
		}
	}
	// change back to original directory
	if err := client.ChangeDir(currentDir); err != nil {
		return errors.New(fmt.Sprintf("Failed to change to folder %s: %s", currentDir, err.Error()))
	}

	// Open and read local file
	file, err := os.Open(fileUpload.LocalFilepath)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to open file %s: %v", fileUpload.LocalFilepath, err))
	}
	defer file.Close()

	path := filepath.Join(fileUpload.FTPFolder, fileUpload.FTPFileName)
	if client.Stor(path, file); err != nil {
		return errors.New(fmt.Sprintf("Failed to upload file: %s", err.Error()))
	}

	return nil
}

func (s *FTP) connect() (*ftp.ServerConn, error) {
	url := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	c, err := ftp.Dial(url, ftp.DialWithTimeout(s.config.Timeout))
	if err != nil {
		return nil, err
	}
	err = c.Login(s.config.User, s.config.Password)
	return c, err
}
