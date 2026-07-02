package sshconn

import (
	"fmt"
	"io"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

type TerminalCallbacks struct {
	OnData func(sessionID string, chunk []byte)
	OnExit func(sessionID string)
}

type termSession struct {
	id      string
	connID  string
	session *ssh.Session
	stdin   io.WriteCloser
	close   sync.Once
}

type Terminal struct {
	connector *Connector

	mu       sync.Mutex
	sessions map[string]*termSession
}

func NewTerminal(connector *Connector) *Terminal {
	return &Terminal{
		connector: connector,
		sessions:  make(map[string]*termSession),
	}
}

type callbackWriter struct {
	sessionID string
	fn        func(string, []byte)
}

func (w callbackWriter) Write(p []byte) (int, error) {
	if w.fn != nil && len(p) > 0 {
		buf := make([]byte, len(p))
		copy(buf, p)
		w.fn(w.sessionID, buf)
	}
	return len(p), nil
}

func (t *Terminal) Open(connID string, cols, rows int, cb TerminalCallbacks) (string, error) {
	client, ok := t.connector.Client(connID)
	if !ok {
		return "", ErrNotConnected
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("%w: new session: %w", ErrTerminalFailed, err)
	}

	cols, rows = normalizeSize(cols, rows)
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		_ = session.Close()
		return "", fmt.Errorf("%w: request pty: %w", ErrTerminalFailed, err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		_ = session.Close()
		return "", fmt.Errorf("%w: stdin pipe: %w", ErrTerminalFailed, err)
	}

	sessionID := uuid.NewString()
	writer := callbackWriter{sessionID: sessionID, fn: cb.OnData}
	session.Stdout = writer
	session.Stderr = writer

	if err := session.Shell(); err != nil {
		_ = session.Close()
		return "", fmt.Errorf("%w: start shell: %w", ErrTerminalFailed, err)
	}

	ts := &termSession{id: sessionID, connID: connID, session: session, stdin: stdin}

	t.mu.Lock()
	t.sessions[sessionID] = ts
	t.mu.Unlock()

	go func() {
		_ = session.Wait()
		t.Close(sessionID)
		if cb.OnExit != nil {
			cb.OnExit(sessionID)
		}
	}()

	return sessionID, nil
}

func (t *Terminal) Write(sessionID string, data []byte) error {
	t.mu.Lock()
	ts, ok := t.sessions[sessionID]
	t.mu.Unlock()
	if !ok {
		return ErrSessionNotFound
	}

	if _, err := ts.stdin.Write(data); err != nil {
		return fmt.Errorf("%w: write: %w", ErrTerminalFailed, err)
	}
	return nil
}

func (t *Terminal) Resize(sessionID string, cols, rows int) error {
	t.mu.Lock()
	ts, ok := t.sessions[sessionID]
	t.mu.Unlock()
	if !ok {
		return ErrSessionNotFound
	}

	cols, rows = normalizeSize(cols, rows)
	if err := ts.session.WindowChange(rows, cols); err != nil {
		return fmt.Errorf("%w: resize: %w", ErrTerminalFailed, err)
	}
	return nil
}

func (t *Terminal) Close(sessionID string) {
	t.mu.Lock()
	ts, ok := t.sessions[sessionID]
	delete(t.sessions, sessionID)
	t.mu.Unlock()
	if ok {
		ts.shutdown()
	}
}

func (t *Terminal) CloseForConnection(connID string) {
	t.mu.Lock()
	var victims []*termSession
	for id, ts := range t.sessions {
		if ts.connID == connID {
			victims = append(victims, ts)
			delete(t.sessions, id)
		}
	}
	t.mu.Unlock()

	for _, ts := range victims {
		ts.shutdown()
	}
}

func (t *Terminal) CloseAll() {
	t.mu.Lock()
	sessions := t.sessions
	t.sessions = make(map[string]*termSession)
	t.mu.Unlock()

	for _, ts := range sessions {
		ts.shutdown()
	}
}

func (ts *termSession) shutdown() {
	ts.close.Do(func() {
		if ts.stdin != nil {
			_ = ts.stdin.Close()
		}
		if ts.session != nil {
			_ = ts.session.Close()
		}
	})
}

func normalizeSize(cols, rows int) (int, int) {
	if cols <= 0 || cols > 1000 {
		cols = 80
	}
	if rows <= 0 || rows > 1000 {
		rows = 24
	}
	return cols, rows
}
