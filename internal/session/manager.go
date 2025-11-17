package session

import (
	"sync"
	"time"
)

// Session armazena dados de rastreamento da conversa.
type Session struct {
	Phone          string
	Name           string
	ConversationID string
	VisitorID      string

	lastActive time.Time
	timer      *time.Timer
}

// Manager controla sessões por número/ticket, expira após ttl.
type Manager struct {
	ttl time.Duration

	mu       sync.Mutex
	sessions map[string]*Session
}

// NewManager cria um gerenciador com ttl configurável.
func NewManager(ttl time.Duration) *Manager {
	return &Manager{
		ttl:      ttl,
		sessions: make(map[string]*Session),
	}
}

// Upsert retorna a sessão do telefone e reseta o timer de expiração.
func (m *Manager) Upsert(phone, name string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[phone]
	if !ok {
		s = &Session{Phone: phone, Name: name}
		m.sessions[phone] = s
	}

	s.Name = name
	s.lastActive = time.Now()

	if s.timer != nil {
		s.timer.Stop()
	}
	s.timer = time.AfterFunc(m.ttl, func() {
		m.expire(phone)
	})

	return s
}

// UpdateConversation grava conversationId/visitorId após a resposta da IA.
func (m *Manager) UpdateConversation(phone, conversationID, visitorID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[phone]; ok {
		s.ConversationID = conversationID
		s.VisitorID = visitorID
	}
}

// expire remove a sessão; chamado via timer.
func (m *Manager) expire(phone string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[phone]; ok {
		if s.timer != nil {
			s.timer.Stop()
		}
		delete(m.sessions, phone)
	}
}

// Get retorna sessão sem alterar o timer.
func (m *Manager) Get(phone string) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[phone]
	return s, ok
}
