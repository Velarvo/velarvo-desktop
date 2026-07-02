package sshconn

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const dialTimeout = 15 * time.Second

type liveConn struct {
	client *ssh.Client
}

type Connector struct {
	knownHostsPath string

	mu    sync.Mutex
	conns map[string]*liveConn
}

func NewConnector(dataDir string) *Connector {
	return &Connector{
		knownHostsPath: filepath.Join(dataDir, "known_hosts"),
		conns:          make(map[string]*liveConn),
	}
}

func (c *Connector) Connect(id string, listing ListingData, password string) error {
	if password == "" {
		return ErrNoPassword
	}

	hostKeyCallback, err := c.hostKeyCallback()
	if err != nil {
		return fmt.Errorf("%w: prepare host key verification: %w", ErrConnectFailed, err)
	}

	config := &ssh.ClientConfig{
		User:            listing.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         dialTimeout,
	}

	target := net.JoinHostPort(listing.Host, strconv.Itoa(normalizePort(listing.Port)))

	live, err := dialDirect(target, config)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConnectFailed, err)
	}

	session, err := live.client.NewSession()
	if err != nil {
		live.close()
		return fmt.Errorf("%w: open session: %w", ErrConnectFailed, err)
	}
	_ = session.Close()

	c.mu.Lock()
	if existing, ok := c.conns[id]; ok {
		existing.close()
	}
	c.conns[id] = live
	c.mu.Unlock()

	return nil
}

func dialDirect(target string, config *ssh.ClientConfig) (*liveConn, error) {
	client, err := ssh.Dial("tcp", target, config)
	if err != nil {
		return nil, err
	}
	return &liveConn{client: client}, nil
}

func (c *Connector) Disconnect(id string) {
	c.mu.Lock()
	live, ok := c.conns[id]
	delete(c.conns, id)
	c.mu.Unlock()

	if ok {
		live.close()
	}
}

func (c *Connector) Client(id string) (*ssh.Client, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	live, ok := c.conns[id]
	if !ok {
		return nil, false
	}
	return live.client, true
}

func (c *Connector) IsConnected(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.conns[id]
	return ok
}

func (c *Connector) ConnectedIDs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	ids := make([]string, 0, len(c.conns))
	for id := range c.conns {
		ids = append(ids, id)
	}
	return ids
}

func (c *Connector) CloseAll() {
	c.mu.Lock()
	conns := c.conns
	c.conns = make(map[string]*liveConn)
	c.mu.Unlock()

	for _, live := range conns {
		live.close()
	}
}

func (l *liveConn) close() {
	if l.client != nil {
		_ = l.client.Close()
	}
}

func (c *Connector) hostKeyCallback() (ssh.HostKeyCallback, error) {
	if err := ensureFile(c.knownHostsPath); err != nil {
		return nil, err
	}

	base, err := knownhosts.New(c.knownHostsPath)
	if err != nil {
		return nil, err
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := base(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			return appendKnownHost(c.knownHostsPath, hostname, remote, key)
		}
		return err
	}, nil
}

func appendKnownHost(path, hostname string, remote net.Addr, key ssh.PublicKey) (err error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // path is app-managed known_hosts storage
	if err != nil {
		return fmt.Errorf("open known_hosts: %w", err)
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()

	addresses := []string{knownhosts.Normalize(hostname)}
	if remoteAddr := knownhosts.Normalize(remote.String()); remoteAddr != addresses[0] {
		addresses = append(addresses, remoteAddr)
	}

	line := knownhosts.Line(addresses, key)
	if _, err := f.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("write known_hosts: %w", err)
	}
	return nil
}

func ensureFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // path is app-managed known_hosts storage
	if err != nil {
		return err
	}
	return f.Close()
}

func normalizePort(port int) int {
	if port <= 0 || port > 65535 {
		return defaultPort
	}
	return port
}
