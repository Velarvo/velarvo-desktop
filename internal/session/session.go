package session

import "sync"

type UserSession struct {
	AccessToken  string
	RefreshToken string
	UserID       string
	Email        string
	FirstName    string
	LastName     string
	DeviceID     string
	MasterKey    []byte
	IsUnlocked   bool
}

type Manager struct {
	mu      sync.RWMutex
	current *UserSession
}

var Default = &Manager{}

func (m *Manager) Set(s *UserSession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.current = s
}

func (m *Manager) IsLoggedIn() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current != nil && m.current.AccessToken != "" && m.current.RefreshToken != ""
}

func (m *Manager) Get() *UserSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

func (m *Manager) UpdateTokens(accessToken, refreshToken string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.current == nil {
		m.current = &UserSession{}
	}

	m.current.AccessToken = accessToken
	m.current.RefreshToken = refreshToken
}

func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current != nil {
		for i := range m.current.MasterKey {
			m.current.MasterKey[i] = 0
		}
		m.current = nil
	}
}
