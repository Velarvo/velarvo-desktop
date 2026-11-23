package sshconn

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const dialTimeout = 15 * time.Second

type Clients struct {
	knownHostsPath string

	mu      sync.Mutex
	clients map[string]*ssh.Client
}

func NewClients(knownHostsPath string) *Clients {
	return &Clients{knownHostsPath: knownHostsPath, clients: make(map[string]*ssh.Client)}
}

func (c *Clients) dial(id, host string, port int, user string, auth []ssh.AuthMethod) (*ssh.Client, error) {
	hostKeyCallback, err := c.hostKeyCallback()
	if err != nil {
		return nil, fmt.Errorf("%w: verify host key: %w", ErrConnectFailed, err)
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)), &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		HostKeyCallback: hostKeyCallback,
		Timeout:         dialTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConnectFailed, err)
	}

	session, err := client.NewSession()
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("%w: open session: %w", ErrConnectFailed, err)
	}
	_ = session.Close()

	c.mu.Lock()
	previous := c.clients[id]
	c.clients[id] = client
	c.mu.Unlock()

	if previous != nil {
		_ = previous.Close()
	}
	return client, nil
}

func (c *Clients) get(id string) (*ssh.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if client, ok := c.clients[id]; ok {
		return client, nil
	}
	return nil, ErrNotConnected
}

func (c *Clients) connected(id string) bool {
	_, err := c.get(id)
	return err == nil
}

func (c *Clients) close(id string) {
	c.mu.Lock()
	client := c.clients[id]
	delete(c.clients, id)
	c.mu.Unlock()

	if client != nil {
		_ = client.Close()
	}
}

func (c *Clients) closeAll() {
	c.mu.Lock()
	clients := c.clients
	c.clients = make(map[string]*ssh.Client)
	c.mu.Unlock()

	for _, client := range clients {
		_ = client.Close()
	}
}

func (c *Clients) hostKeyCallback() (ssh.HostKeyCallback, error) {
	file, err := os.OpenFile(c.knownHostsPath, os.O_CREATE|os.O_RDONLY, 0o600)
	if err != nil {
		return nil, err
	}
	_ = file.Close()

	known, err := knownhosts.New(c.knownHostsPath)
	if err != nil {
		return nil, err
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := known(hostname, remote, key)
		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			return c.trustHost(hostname, remote, key)
		}
		return err
	}, nil
}

func (c *Clients) trustHost(hostname string, remote net.Addr, key ssh.PublicKey) (err error) {
	file, err := os.OpenFile(c.knownHostsPath, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open known_hosts: %w", err)
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	addresses := []string{knownhosts.Normalize(hostname)}
	if addr := knownhosts.Normalize(remote.String()); addr != addresses[0] {
		addresses = append(addresses, addr)
	}
	if _, err := file.WriteString(knownhosts.Line(addresses, key) + "\n"); err != nil {
		return fmt.Errorf("write known_hosts: %w", err)
	}
	return nil
}
