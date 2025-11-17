package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gothout/goqueue" // TODO: conferir nome do módulo e import path real

	"whatsapp-ia-integrator/internal/csa"
)

// OutboxJob é o que vai para a fila.
type OutboxJob struct {
	Phone          string // número do cliente
	ConversationID string // conversationId do Chatvolt (ticket)
	Text           string
}

// Outbox encapsula a fila + workers que consomem e enviam pra CSA.
type Outbox struct {
	wg      sync.WaitGroup
	csa     *csa.Client
	workers int

	// fila goqueue
	queue *goqueue.Queue // TODO: trocar pelo tipo certo da tua lib (genérico, interface, etc.)

	shutdown chan struct{}
}

// NewOutbox cria a fila e prepara os workers.
// buffer aqui é ignorado porque quem manda no buffer é a implementação da goqueue.
func NewOutbox(csaClient *csa.Client, workers int) *Outbox {
	if workers <= 0 {
		workers = 3
	}

	// TODO: ajustar para a API de criação da fila:
	// algo como:
	//   q := goqueue.New()        // se for ilimitada
	//   q := goqueue.New(100)     // se precisar informar capacidade
	q := goqueue.New() // PLACEHOLDER – AJUSTAR

	return &Outbox{
		csa:      csaClient,
		workers:  workers,
		queue:    q,
		shutdown: make(chan struct{}),
	}
}

// Start sobe os workers que vão ficar dando Get() / Pop() na fila.
func (o *Outbox) Start() {
	for i := 0; i < o.workers; i++ {
		o.wg.Add(1)
		go o.worker(i + 1)
	}
}

// worker é quem consome da fila e chama a CSA.
func (o *Outbox) worker(id int) {
	defer o.wg.Done()

	for {
		select {
		case <-o.shutdown:
			log.Printf("[outbox-worker-%d] encerrando", id)
			return
		default:
			// TODO: ajustar para a API real do goqueue.
			// Exemplo se for algo tipo:
			//   item := o.queue.Get()
			//   job := item.(OutboxJob)
			item := o.queue.Pop() // PLACEHOLDER – AJUSTAR NOME/ASSINATURA
			if item == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			job, ok := item.(OutboxJob)
			if !ok {
				log.Printf("[outbox-worker-%d] item de tipo inesperado: %#v", id, item)
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			err := o.csa.SendMessage(ctx, &csa.SendMessageRequest{
				Destination: job.Phone,
				Text:        job.Text,
				Type:        "text",
				Preview:     true,
			// InstanceID/Product/Provider/Name vêm por default no c
