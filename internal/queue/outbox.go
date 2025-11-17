package queue

import (
	"context"
	"log"
	"sync"
	"time"

	goqueue "github.com/gothout/goqueue"

	"whatsapp-ia-integrator/internal/csa"
)

// OutboxJob é o item colocado na fila para envio à CSA.
type OutboxJob struct {
	Phone          string
	ConversationID string
	Text           string
}

// Outbox é uma fila de saída com workers concorrentes, apoiada pelo goqueue para armazenar mensagens.
type Outbox struct {
	csa     *csa.Client
	workers int

	queue    *goqueue.Queue[OutboxJob]
	notify   chan struct{}
	wg       sync.WaitGroup
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
		queue:    goqueue.NewQueue[OutboxJob](100),
		notify:   make(chan struct{}, 1),
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
	close(o.notify)
	o.wg.Wait()
}

// Enqueue adiciona um job na fila.
func (o *Outbox) Enqueue(job OutboxJob) {
	if ok := o.queue.Enqueue(job); !ok {
		log.Printf("[outbox] fila cheia, descartando mensagem para %s", job.Phone)
		return
	}

	select {
	case o.notify <- struct{}{}:
	default:
	}
}

func (o *Outbox) worker(id int) {
	defer o.wg.Done()

	for {
		select {
		case <-o.shutdown:
			log.Printf("[outbox-worker-%d] encerrando", id)
			return
		case _, ok := <-o.notify:
			if !ok {
				log.Printf("[outbox-worker-%d] canal fechado", id)
				return
			}

			for {
				job, ok := o.queue.Dequeue()
				if !ok {
					break
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
}
