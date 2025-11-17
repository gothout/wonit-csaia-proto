package session

import (
	"log"
	"sync"
	"time"

	"github.com/gothout/goqueue"
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

// SessionEvent representa um estágio do ciclo de vida da sessão para logging.
type SessionEvent struct {
	Phone        string
	Name         string
	Conversation string
	Visitor      string
	Stage        string
	Timestamp    time.Time
}

// Manager controla sessões por número/ticket, expira após ttl.
type Manager struct {
	ttl time.Duration

	mu       sync.Mutex
	sessions map[string]*Session

	events   *goqueue.Queue[SessionEvent]
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewManager cria um gerenciador com ttl configurável.
func NewManager(ttl time.Duration) *Manager {
	m := &Manager{
		ttl:      ttl,
		sessions: make(map[string]*Session),
		events:   goqueue.NewQueue[SessionEvent](0),
		stopCh:   make(chan struct{}),
	}

	go m.processEvents()

	return m
}

// Upsert retorna a sessão do telefone e reseta o timer de expiração.
func (m *Manager) Upsert(phone, name string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[phone]
	stage := "upsert:existing"
	if !ok {
		s = &Session{Phone: phone, Name: name}
		m.sessions[phone] = s
		stage = "upsert:new"
	}

	s.Name = name
	s.lastActive = time.Now()

	if s.timer != nil {
		s.timer.Stop()
	}
	s.timer = time.AfterFunc(m.ttl, func() {
		m.expire(phone)
	})

	m.recordStage(s, stage)

	return s
}

// UpdateConversation grava conversationId/visitorId após a resposta da IA.
func (m *Manager) UpdateConversation(phone, conversationID, visitorID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[phone]; ok {
		s.ConversationID = conversationID
		s.VisitorID = visitorID
		m.recordStage(s, "conversation:update")
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
		m.recordStage(s, "expired")
		delete(m.sessions, phone)
	}
}

// Get retorna sessão sem alterar o timer.
func (m *Manager) Get(phone string) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[phone]
	if ok {
		m.recordStage(s, "get")
	}
	return s, ok
}

// Stop encerra o processamento de eventos de sessão.
func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopCh)
	})
}

func (m *Manager) recordStage(s *Session, stage string) {
	if s == nil {
		return
	}

	_ = m.events.Enqueue(SessionEvent{
		Phone:        s.Phone,
		Name:         s.Name,
		Conversation: s.ConversationID,
		Visitor:      s.VisitorID,
		Stage:        stage,
		Timestamp:    time.Now(),
	})
}

func (m *Manager) processEvents() {
	for {
		select {
		case <-m.stopCh:
			return
		default:
		}

		if evt, ok := m.events.Dequeue(); ok {
			log.Printf("[session] stage=%s phone=%s name=%s conversation=%s visitor=%s at=%s", evt.Stage, evt.Phone, evt.Name, evt.Conversation, evt.Visitor, evt.Timestamp.Format(time.RFC3339))
			continue
		}

		time.Sleep(25 * time.Millisecond)
	}
}
