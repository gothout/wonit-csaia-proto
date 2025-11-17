package model

// InboundWebhook representa o payload recebido do CSA.
type InboundWebhook struct {
	Event     string         `json:"event,omitempty"`
	Timestamp int64          `json:"timestamp,omitempty"`
	Message   InboundMessage `json:"message"`
	Contact   Contact        `json:"contact"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// InboundMessage traz o conteúdo textual ou multimídia.
type InboundMessage struct {
	ID        string `json:"id,omitempty"`
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Caption   string `json:"caption,omitempty"`
	URL       string `json:"url,omitempty"`
	Filename  string `json:"filename,omitempty"`
	Latitude  string `json:"latitude,omitempty"`
	Longitude string `json:"longitude,omitempty"`
	// Alguns provedores enviam o número de origem diretamente aqui.
	From string `json:"from,omitempty"`
}

// Contact representa quem enviou a mensagem.
type Contact struct {
	Phone string `json:"phone"`
	Name  string `json:"name,omitempty"`
}
