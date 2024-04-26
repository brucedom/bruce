package rssh

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type RSSH struct {
	Host   string
	User   string
	Key    []byte
	Port   string
	client *ssh.Client
}

func NewRSSH(host, user, privkey string) (*RSSH, error) {
	if privkey == "" {
		privkey = os.ExpandEnv("$HOME/.ssh/id_rsa")
	}
	keyBytes, err := os.ReadFile(privkey)
	if err != nil {
		log.Error().Err(err).Msg("failed to read private key")
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	port := ":22"
	if strings.Contains(host, ":") {
		splits := strings.Split(host, ":")
		host = splits[0]
		port = ":" + splits[1]
	}
	rsshc := &RSSH{
		Host: fmt.Sprintf("[%s]", host), // Ensure IPv6 addresses are correctly formatted
		User: user,
		Key:  keyBytes,
		Port: port,
	}
	err = rsshc.setup()
	if err != nil {
		return nil, err
	}
	return rsshc, nil
}

func (r *RSSH) setup() error {
	signer, err := ssh.ParsePrivateKey(r.Key)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	khPath := os.ExpandEnv("$HOME/.ssh/known_hosts")
	hkCback, err := knownhosts.New(khPath)
	if err != nil {
		return fmt.Errorf("failed to create known_hosts callback: %w", err)
	}

	conf := &ssh.ClientConfig{
		User: r.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hkCback,
	}

	addr := fmt.Sprintf("%s%s", r.Host, r.Port)
	client, err := ssh.Dial("tcp", addr, conf)
	if err != nil {
		return fmt.Errorf("failed to connect to host %s: %w", addr, err)
	}
	r.client = client
	return nil
}

func (r *RSSH) ExecCommand(cmd string) (string, error) {
	session, err := r.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer log.Error().Err(session.Close())

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("command execution failed: %w | stderr: %s", err, stderr.String())
	}

	if stderr.Len() > 0 {
		return "", fmt.Errorf("stderr: %s", stderr.String())
	}

	return stdout.String(), nil
}

func (r *RSSH) Close() {
	if r.client != nil {
		log.Error().Err(r.client.Close())
		r.client = nil
	}
}
