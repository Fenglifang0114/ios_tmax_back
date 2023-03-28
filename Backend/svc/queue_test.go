package svc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	size := 100
	q := NewCircularBuffer(size)

	for i := 0; i < size; i++ {
		assert.NoError(t, q.Enqueue(byte(i)))
	}

	data, _ := q.Peekn(50, 50)
	fmt.Printf("%v\n", data)
	for i := 0; i < 50; i++ {
		assert.Equal(t, byte(i+50), data[i])
	}

	// can't insert new data.
	assert.Error(t, q.Enqueue(0))
	assert.Equal(t, errFull, q.Enqueue(0))

	assert.Equal(t, size, q.DataLen())
	for i := 0; i < size; i++ {
		v, _ := q.Peek(i)
		assert.Equal(t, byte(i), v)
	}

	for i := 0; i < size/2; i++ {
		v, err := q.Dequeue()
		assert.Equal(t, byte(i), v)
		assert.NoError(t, err)
	}

	for i := 0; i < size/2; i++ {
		assert.NoError(t, q.Enqueue(byte(i*3)))
	}

	assert.Equal(t, true, q.IsFull())
	data, _ = q.DequeueN(size)
	fmt.Printf("%v\n", data)
	assert.Equal(t, true, q.IsEmpty())

	// TODO: add test GetDataLen()
	// no task
	// _, err := q.Dequeue()
	// assert.Error(t, err)
	// assert.Equal(t, errNoTask, err)
}

func BenchmarkCircularBufferEnqueueDequeue(b *testing.B) {
	q := NewCircularBuffer(b.N)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Enqueue(byte(i))
		_, _ = q.Dequeue()
	}
}

func BenchmarkCircularBufferEnqueue(b *testing.B) {
	q := NewCircularBuffer(b.N)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Enqueue(byte(i))
	}
}

func BenchmarkCircularBufferDequeue(b *testing.B) {
	q := NewCircularBuffer(b.N)

	for i := 0; i < b.N; i++ {
		_ = q.Enqueue(byte(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = q.Dequeue()
	}
}
