package go4ftp

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
	client *ftp.ServerConn
}

func newFTP(config ConnConfig) Instance {
	return &FTP{config, nil}
}

func (s *FTP) Connect() error {
	url := fmt.Sprintf("%v:%v", s.config.Host, s.config.Port)
	c, err := ftp.Dial(url, ftp.DialWithTimeout(s.config.Timeout))
	if err != nil {
		return err
	}
	err = c.Login(s.config.User, s.config.Password)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to connect to server: %s", err.Error()))
	}
	s.client = c
	return nil
}

// Close closes the connection to the FTP server.
func (s *FTP) Close() error {
	if s.client != nil {
		return s.client.Quit()
	}
	return nil
}

func (s *FTP) Ping() error {
	err := s.Connect()
	if err != nil {
		return err
	}
	return s.Close()
}

func (s *FTP) UploadFile(fileUpload FileUpload) error {
	// Recursively create folder in FTP server
	folders := strings.Split(fileUpload.FTPFolder, "/")
	currentDir, _ := s.client.CurrentDir()
	for _, folder := range folders {

		current, _ := s.client.CurrentDir()
		current = filepath.Join(current, folder)

		if err := s.client.ChangeDir(current); err != nil {
			// create folder
			if err := s.client.MakeDir(folder); err != nil {
				return errors.New(fmt.Sprintf("Failed to create folder %s: %s", folder, err.Error()))
			}
		}
		// change to folder
		if err := s.client.ChangeDir(current); err != nil {
			return errors.New(fmt.Sprintf("Failed to change to folder %s: %s", folder, err.Error()))
		}
	}
	// change back to original directory
	if err := s.client.ChangeDir(currentDir); err != nil {
		return errors.New(fmt.Sprintf("Failed to change to folder %s: %s", currentDir, err.Error()))
	}

	// Open and read local file
	file, err := os.Open(fileUpload.LocalFilepath)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to open file %s: %v", fileUpload.LocalFilepath, err))
	}
	defer file.Close()

	path := filepath.Join(fileUpload.FTPFolder, fileUpload.FTPFileName)
	if s.client.Stor(path, file); err != nil {
		return errors.New(fmt.Sprintf("Failed to upload file: %s", err.Error()))
	}

	return nil
}

func (s *FTP) Read(path string) ([]Entries, error) {

	err := s.Connect()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	result := make([]Entries, 0)
	entries, err := s.client.List(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		result = append(result, Entries{
			Name: entry.Name,
			Size: entry.Size,
		})
	}

	return result, nil
}
