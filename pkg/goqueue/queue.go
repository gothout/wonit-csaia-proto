package goqueue

import "sync"

// Queue é uma fila thread-safe simples baseada em slice.
type Queue[T any] struct {
	mu       sync.Mutex
	items    []T
	capacity int
}

// NewQueue cria uma nova fila com a capacidade desejada (0 ou negativa significa sem limite).
func NewQueue[T any](capacity int) *Queue[T] {
	return &Queue[T]{capacity: capacity}
}

// Enqueue adiciona um item à fila, retornando false se estiver cheia.
func (q *Queue[T]) Enqueue(item T) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.capacity > 0 && len(q.items) >= q.capacity {
		return false
	}

	q.items = append(q.items, item)
	return true
}

// Dequeue remove e retorna o próximo item. O bool indica se havia elemento.
func (q *Queue[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var zero T
	if len(q.items) == 0 {
		return zero, false
	}

	item := q.items[0]
	q.items[0] = zero
	q.items = q.items[1:]
	return item, true
}

// Len retorna o tamanho atual da fila.
func (q *Queue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}
