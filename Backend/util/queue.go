package util

import (
	"errors"
	"sync"

	"tmaxsrv/log"
)

// type T interface{}

var (
	ErrFull          = errors.New("full")
	ErrNoTask        = errors.New("no task")
	ErrNotEnoughTask = errors.New("not enough task")
)

type CircularBuffer struct {
	sync.Mutex
	taskQueue []byte
	Capacity  int
	head      int
	tail      int
	full      bool
}

func (s *CircularBuffer) IsEmpty() bool {
	return s.head == s.tail && !s.full
}

func (s *CircularBuffer) IsFull() bool {
	return s.full
}

func (s *CircularBuffer) Enqueue(task byte) error {
	s.Lock()
	defer s.Unlock()

	if s.IsFull() {
		log.Log.Error("queue full\n")
		return ErrFull
	}

	s.taskQueue[s.tail] = task
	s.tail = (s.tail + 1) % s.Capacity
	s.full = s.head == s.tail

	return nil
}

// if tasks is too long to enqueue then return an full error and none of the task will be enqueued
func (s *CircularBuffer) EnqueueN(tasks []byte, len int) error {
	s.Lock()
	defer s.Unlock()

	if s.dataLen()+len > s.Capacity {
		log.Log.Warn("queue full")
		return ErrFull
	}

	for i := 0; i < len; i++ {
		s.taskQueue[s.tail] = tasks[i]
		s.tail = (s.tail + 1) % s.Capacity
		// should not full
	}

	return nil
}

func (s *CircularBuffer) Reset() {
	s.head = 0
	s.tail = 0
	s.full = false
}

func (s *CircularBuffer) Dequeue() (byte, error) {
	s.Lock()
	defer s.Unlock()

	if s.IsEmpty() {
		return 0, ErrNoTask
	}

	data := s.taskQueue[s.head]
	s.full = false
	s.head = (s.head + 1) % s.Capacity

	return data, nil
}

func (s *CircularBuffer) DequeueN(size int) ([]byte, error) {
	s.Lock()
	defer s.Unlock()

	if s.IsEmpty() {
		return nil, ErrNoTask
	}

	len := s.dataLen()
	if len < size {
		return nil, ErrNotEnoughTask
	}

	result := []byte{}
	idx := s.head
	s.full = false
	for i := 0; i < size; i++ {
		result = append(result, s.taskQueue[idx])
		idx = (idx + 1) % s.Capacity
	}
	s.head = (s.head + size) % s.Capacity

	return result, nil
}

func NewCircularBuffer(size int) *CircularBuffer {
	w := &CircularBuffer{
		taskQueue: make([]byte, size),
		Capacity:  size,
	}

	return w
}

func (s *CircularBuffer) Peek(i int) (byte, error) {
	s.Lock()
	defer s.Unlock()

	if s.IsEmpty() {
		return 0, ErrNoTask
	}
	offset := (s.head + i) % s.Capacity
	data := s.taskQueue[offset]

	return data, nil
}

func (s *CircularBuffer) GetDataLen() int {
	s.Lock()
	defer s.Unlock()
	return s.dataLen()
}

func (s *CircularBuffer) dataLen() int {
	if s.IsEmpty() {
		return 0
	}

	if s.IsFull() {
		return s.Capacity
	}

	if s.tail >= s.head {
		return s.tail - s.head
	}

	if s.head > s.tail {
		return s.Capacity - s.head + s.tail
	}

	return 0 // should never happen
}

func (s *CircularBuffer) Peekn(offset int, size int) ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	if s.IsEmpty() {
		return nil, ErrNoTask
	}

	if offset+size > s.dataLen() {
		return nil, ErrNotEnoughTask
	}

	result := []byte{}
	idx := (s.head + offset) % s.Capacity
	for i := 0; i < size; i++ {
		result = append(result, s.taskQueue[idx])
		idx = (idx + 1) % s.Capacity
	}

	return result, nil
}

func (s *CircularBuffer) PeekAll() []byte {
	s.Lock()
	defer s.Unlock()

	dataLen := s.dataLen()
	result := []byte{}
	idx := s.head
	for i := 0; i < dataLen; i++ {
		result = append(result, s.taskQueue[idx])
		idx = (idx + 1) % s.Capacity
	}

	return result
}
