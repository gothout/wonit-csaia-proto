package whatsapp

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"whatsapp-ia-integrator/internal/chatvolt"
	"whatsapp-ia-integrator/internal/model"
	"whatsapp-ia-integrator/internal/queue"
	"whatsapp-ia-integrator/internal/session"
)

// Handler recebe webhooks da CSA e orquestra o fluxo IA -> CSA.
type Handler struct {
	chatvolt *chatvolt.Client
	sessions *session.Manager
	outbox   *queue.Outbox
}

func NewHandler(cv *chatvolt.Client, sm *session.Manager, out *queue.Outbox) *Handler {
	return &Handler{chatvolt: cv, sessions: sm, outbox: out}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload model.InboundWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("[webhook] erro lendo payload: %v", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	phone := strings.TrimSpace(payload.From)
	if phone == "" {
		phone = strings.TrimSpace(payload.PhoneFromRaw())
	}
	if phone == "" && payload.RawContact != nil {
		phone = strings.TrimSpace(payload.RawContact["phone"])
	}
	if phone == "" {
		log.Printf("[webhook] payload sem telefone: %#v", payload)
		http.Error(w, "missing phone", http.StatusBadRequest)
		return
	}

	text := strings.TrimSpace(payload.MessageText)
	if text == "" {
		text = strings.TrimSpace(payload.TextFromRaw())
	}
	if text == "" {
		log.Printf("[webhook] mensagem sem texto ignorada: %#v", payload)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	name := ""
	if payload.RawContact != nil {
		name = payload.RawContact["name"]
	}
	sess := h.sessions.Upsert(phone, name)

	req := chatvolt.QueryRequest{
		Query:          text,
		ConversationID: sess.ConversationID,
		VisitorID:      sess.VisitorID,
		Contact: &chatvolt.Contact{
			FirstName: name,
			Phone:     phone,
		},
	}

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	resp, err := h.chatvolt.Query(ctx, req)
	if err != nil {
		log.Printf("[webhook] erro chamando chatvolt: %v", err)
		http.Error(w, "failed to query IA", http.StatusBadGateway)
		return
	}

	h.sessions.UpdateConversation(phone, resp.ConversationID, resp.VisitorID)

	h.outbox.Enqueue(queue.OutboxJob{
		Phone:          phone,
		ConversationID: resp.ConversationID,
		Text:           resp.Answer,
	})

	w.WriteHeader(http.StatusAccepted)
}
