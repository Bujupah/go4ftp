package go4ftp

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
	"strings"
)

type SFTP struct {
	config    ConnConfig
	sshClient *ssh.Client
	client    *sftp.Client
}

func NewSFTP(config ConnConfig) Instance {
	return &SFTP{config, nil, nil}

}

func (s *SFTP) Ping() error {
	err := s.Connect()
	if err != nil {
		return err
	}
	defer s.Close()
	return nil
}

func (s *SFTP) UploadFile(fileUpload FileUpload) error {
	if err := s.client.MkdirAll(fileUpload.FTPFolder); err != nil {
		return err
	}

	// Create file in SFTP server
	path := filepath.Join(fileUpload.FTPFolder, fileUpload.FTPFileName)
	f, err := s.client.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Open and read local file
	file, err := os.ReadFile(fileUpload.LocalFilepath)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to open file %s: %v", fileUpload.LocalFilepath, err))
	}

	// Write local file to SFTP server
	if _, err := f.Write(file); err != nil {
		return errors.New(fmt.Sprintf("Failed to upload file: %s", err.Error()))
	}
	f.Close()

	return nil
}

func (s *SFTP) Connect() error {
	config, err := sshClientConfig(s.config)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	s.sshClient, err = ssh.Dial("tcp", url, config)
	if err != nil {
		return err
	}

	client, err := sftp.NewClient(s.sshClient)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

func (s *SFTP) Close() error {
	if s.sshClient != nil {
		defer s.sshClient.Close()
	}
	if s.client != nil {
		defer s.client.Close()
	}
	return nil
}

func sshClientConfig(conn ConnConfig) (*ssh.ClientConfig, error) {
	hostKeyCallback := ssh.InsecureIgnoreHostKey()

	if conn.IgnoreHostKey == false {
		hostKey, err := getHostKey(conn.Host)
		if err != nil {
			return nil, err
		}
		hostKeyCallback = ssh.FixedHostKey(*hostKey)
	}

	return &ssh.ClientConfig{
		User: conn.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(conn.Password),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         conn.Timeout,
	}, nil
}

func getHostKey(host string) (*ssh.PublicKey, error) {
	// parse OpenSSH known_hosts file
	// ssh or use ssh-keyscan to get initial key
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Error parsing %q: %v", fields[2], err))
			}
			break
		}
	}

	if hostKey == nil {
		return nil, errors.New(fmt.Sprintf("No hostkey found for %s", host))
	}

	return &hostKey, nil
}
