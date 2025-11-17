package queue

import (
	"strings"
	"sync"
	"time"
)

// JobStatus representa o estado conhecido de um envio na CSA.
type JobStatus string

const (
	JobStatusSubmitted JobStatus = "submitted"
	JobStatusSent      JobStatus = "sent"
	JobStatusDelivered JobStatus = "delivered"
	JobStatusFailed    JobStatus = "failed"
	JobStatusPending   JobStatus = "pending"
	JobStatusEnqueued  JobStatus = "enqueued"
)

// JobInfo agrega metadados de rastreamento.
type JobInfo struct {
	MessageID      string    `json:"messageId"`
	Phone          string    `json:"phone,omitempty"`
	ConversationID string    `json:"conversationId,omitempty"`
	Status         JobStatus `json:"status"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// JobManager mantém o mapa de status por messageId.
type JobManager struct {
	mu   sync.RWMutex
	jobs map[string]JobInfo
}

func NewJobManager() *JobManager {
	return &JobManager{jobs: make(map[string]JobInfo)}
}

// UpsertStatus cria ou atualiza o status de um job.
func (m *JobManager) UpsertStatus(messageID string, status JobStatus, phone, conversationID string) JobInfo {
	if messageID == "" {
		return JobInfo{}
	}

	normalized := JobStatus(strings.ToLower(string(status)))

	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.jobs[messageID]
	if !ok {
		info.MessageID = messageID
	}

	if info.Phone == "" {
		info.Phone = phone
	}
	if info.ConversationID == "" {
		info.ConversationID = conversationID
	}
	if normalized != "" {
		info.Status = normalized
	}
	info.UpdatedAt = time.Now().UTC()

	m.jobs[messageID] = info
	return info
}

// Get retorna informações de um job específico.
func (m *JobManager) Get(messageID string) (JobInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, ok := m.jobs[messageID]
	return info, ok
}
