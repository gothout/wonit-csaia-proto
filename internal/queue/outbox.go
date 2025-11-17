package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"whatsapp-ia-integrator/internal/csa"
)

// OutboxJob é o item colocado na fila para envio à CSA.
type OutboxJob struct {
	Phone          string
	ConversationID string
	Text           string
}

// Outbox é uma fila simples de saída com workers concorrentes.
type Outbox struct {
	csa     *csa.Client
	workers int

	jobs chan OutboxJob
	wg   sync.WaitGroup

	shutdown chan struct{}
}

// NewOutbox cria a fila e inicializa os canais.
func NewOutbox(csaClient *csa.Client, workers int) *Outbox {
	if workers <= 0 {
		workers = 3
	}

	return &Outbox{
		csa:      csaClient,
		workers:  workers,
		jobs:     make(chan OutboxJob, 100),
		shutdown: make(chan struct{}),
	}
}

// Start dispara os workers de consumo.
func (o *Outbox) Start() {
	for i := 0; i < o.workers; i++ {
		o.wg.Add(1)
		go o.worker(i + 1)
	}
}

// Stop sinaliza shutdown e aguarda workers terminarem.
func (o *Outbox) Stop() {
	close(o.shutdown)
	close(o.jobs)
	o.wg.Wait()
}

// Enqueue adiciona um job na fila.
func (o *Outbox) Enqueue(job OutboxJob) {
	select {
	case o.jobs <- job:
	default:
		log.Printf("[outbox] fila cheia, descartando mensagem para %s", job.Phone)
	}
}

func (o *Outbox) worker(id int) {
	defer o.wg.Done()

	for {
		select {
		case <-o.shutdown:
			log.Printf("[outbox-worker-%d] encerrando", id)
			return
		case job, ok := <-o.jobs:
			if !ok {
				log.Printf("[outbox-worker-%d] canal fechado", id)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			err := o.csa.SendMessage(ctx, &csa.SendMessageRequest{
				Destination: job.Phone,
				Text:        job.Text,
				Type:        "text",
				Preview:     true,
			})
			cancel()

			if err != nil {
				log.Printf("[outbox-worker-%d] erro enviando para %s (conv %s): %v", id, job.Phone, job.ConversationID, err)
				continue
			}

			log.Printf("[outbox-worker-%d] mensagem enviada para %s (conv %s)", id, job.Phone, job.ConversationID)
		}
	}
}
