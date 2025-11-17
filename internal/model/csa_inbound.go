package model

// InboundWebhook representa o payload recebido do CSA/Gupshup.
type InboundWebhook struct {
	Channel           string            `json:"channel"`
	Event             string            `json:"event"`
	MessageID         string            `json:"messageId"`
	PlatformMessageID string            `json:"platformMessageId"`
	PlatformID        string            `json:"platformId"`
	From              string            `json:"from"`
	To                string            `json:"to"`
	Type              string            `json:"type"`
	Status            string            `json:"status"`
	MessageText       string            `json:"messageText"`
	ConversationID    string            `json:"conversationId"`
	Timestamp         int64             `json:"timestamp"`
	RawPayload        map[string]any    `json:"rawPayload,omitempty"`
	RawContact        map[string]string `json:"contact,omitempty"`
}

// TextFromRaw tenta extrair o corpo de texto do rawPayload.
func (i InboundWebhook) TextFromRaw() string {
	if i.RawPayload == nil {
		return ""
	}

	if text, ok := i.RawPayload["text"].(map[string]any); ok {
		if body, ok := text["body"].(string); ok {
			return body
		}
	}

	if caption, ok := i.RawPayload["caption"].(string); ok {
		return caption
	}

	return ""
}

// PhoneFromRaw retorna um telefone de fallback quando o campo From não vem preenchido.
func (i InboundWebhook) PhoneFromRaw() string {
	if i.RawPayload == nil {
		return ""
	}

	if from, ok := i.RawPayload["from"].(string); ok {
		return from
	}

	return ""
}
