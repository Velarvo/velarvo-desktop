package sshconn

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/Velarvo/velarvo-desktop/internal/apperr"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/crypto/ssh"
)

var (
	ErrTerminalFailed  = apperr.New("SSH_TERMINAL_FAILED", "ssh terminal failed")
	ErrSessionNotFound = apperr.New("SSH_TERMINAL_FAILED", "ssh terminal session not found")

	ErrSessionLimit = apperr.New("SSH_SESSION_LIMIT", "ssh server refused another session")
)

const (
	EventTerminalData = "ssh:terminal:data"
	EventTerminalExit = "ssh:terminal:exit"
)

type TerminalData struct {
	SessionID string `json:"sessionId"`
	Data      []byte `json:"data"`
}

type TerminalExit struct {
	SessionID string `json:"sessionId"`
}

func init() {
	application.RegisterEvent[TerminalData](EventTerminalData)
	application.RegisterEvent[TerminalExit](EventTerminalExit)
}

type session struct {
	ssh   *ssh.Session
	stdin io.WriteCloser
}

type Terminal struct {
	clients *Clients

	mu       sync.Mutex
	sessions map[string]*session
}

func NewTerminal(clients *Clients) *Terminal {
	return &Terminal{clients: clients, sessions: make(map[string]*session)}
}

func (t *Terminal) Open(connectionID string, cols, rows int) (string, error) {
	client, err := t.clients.get(connectionID)
	if err != nil {
		return "", err
	}

	sshSession, err := client.NewSession()
	if err != nil {
		var openErr *ssh.OpenChannelError
		if errors.As(err, &openErr) && openErr.Reason == ssh.Prohibited {
			return "", fmt.Errorf("%w: %w", ErrSessionLimit, err)
		}
		return "", fmt.Errorf("%w: %w", ErrTerminalFailed, err)
	}
	fail := func(step string, err error) (string, error) {
		_ = sshSession.Close()
		return "", fmt.Errorf("%w: %s: %w", ErrTerminalFailed, step, err)
	}

	cols, rows = clampSize(cols, rows)
	modes := ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}
	if err := sshSession.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		return fail("request pty", err)
	}
	stdin, err := sshSession.StdinPipe()
	if err != nil {
		return fail("open stdin", err)
	}

	id := uuid.NewString()
	output := outputWriter(id)
	sshSession.Stdout, sshSession.Stderr = output, output
	if err := sshSession.Shell(); err != nil {
		return fail("start shell", err)
	}

	t.mu.Lock()
	t.sessions[id] = &session{ssh: sshSession, stdin: stdin}
	t.mu.Unlock()

	go func() {
		_ = sshSession.Wait()
		t.Close(id)
		emit(EventTerminalExit, TerminalExit{SessionID: id})
	}()
	return id, nil
}

func (t *Terminal) Write(sessionID, data string) error {
	s, err := t.session(sessionID)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(s.stdin, data); err != nil {
		return fmt.Errorf("%w: write: %w", ErrTerminalFailed, err)
	}
	return nil
}

func (t *Terminal) Resize(sessionID string, cols, rows int) error {
	s, err := t.session(sessionID)
	if err != nil {
		return err
	}
	cols, rows = clampSize(cols, rows)
	if err := s.ssh.WindowChange(rows, cols); err != nil {
		return fmt.Errorf("%w: resize: %w", ErrTerminalFailed, err)
	}
	return nil
}

func (t *Terminal) Close(sessionID string) {
	t.mu.Lock()
	s := t.sessions[sessionID]
	delete(t.sessions, sessionID)
	t.mu.Unlock()

	if s != nil {
		_ = s.stdin.Close()
		_ = s.ssh.Close()
	}
}

func (t *Terminal) CloseAll() {
	t.mu.Lock()
	sessions := t.sessions
	t.sessions = make(map[string]*session)
	t.mu.Unlock()

	for _, s := range sessions {
		_ = s.stdin.Close()
		_ = s.ssh.Close()
	}
}

func (t *Terminal) session(id string) (*session, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s, ok := t.sessions[id]; ok {
		return s, nil
	}
	return nil, ErrSessionNotFound
}

type outputWriter string

func (id outputWriter) Write(p []byte) (int, error) {
	if len(p) > 0 {
		emit(EventTerminalData, TerminalData{SessionID: string(id), Data: append([]byte(nil), p...)})
	}
	return len(p), nil
}

func clampSize(cols, rows int) (int, int) {
	if cols <= 0 || cols > 1000 {
		cols = 80
	}
	if rows <= 0 || rows > 1000 {
		rows = 24
	}
	return cols, rows
}

func emit(name string, data any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data)
	}
}
