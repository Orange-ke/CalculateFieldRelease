package deque

import (
	"fmt"
	"lz/model"
	"testing"
	"time"
)

func TestArrDeque_Traverse(t *testing.T) {
	deque := NewArrDeque(4000)
	for i := 0; i < 4000; i++ {
		deque.AddFirst(1550.0)
	}
	start := time.Now()
	for c := 0; c < 100; c++ {
		deque.Traverse(func(z int, item *model.ItemType) {
			for i := 0; i < len(item); i++ {
				for j := 0; j < len(item[0]); j++ {
					item[i][j] += 1
				}
			}
		}, 0, 0)
	}

	fmt.Println(time.Since(start))
}

func BenchmarkArrDeque_AddFirst(b *testing.B) {
	deque := NewArrDeque(4000)
	for i := 0; i < b.N; i++ {
		deque.AddFirst(1000)
		deque.RemoveFirst()
	}
}

func BenchmarkArrDeque_RemoveLast(b *testing.B) {
	deque := NewArrDeque(4000)
	for i := 0; i < b.N; i++ {
		deque.AddLast(1000)
		deque.RemoveLast()
	}
}

func TestArrDeque_Funcs(t *testing.T) {
	deque := NewArrDeque(4000)
	for i := 0; i < 4000; i++ {
		deque.AddFirst(1550.0)
	}
	fmt.Println(deque.container.start, deque.container.end, deque.container1.start, deque.container1.end, deque.container.isFull, deque.container1.isFull, deque.isFull, deque.state)
	deque.RemoveLast()
	fmt.Println(deque.container.start, deque.container.end, deque.container1.start, deque.container1.end, deque.container.isFull, deque.container1.isFull, deque.isFull, deque.state)
	deque.AddFirst(1550.0)
	fmt.Println(deque.container.start, deque.container.end, deque.container1.start, deque.container1.end, deque.container.isFull, deque.container1.isFull, deque.isFull, deque.state)
	deque.AddFirst(1550.0)
	fmt.Println(deque.container.start, deque.container.end, deque.container1.start, deque.container1.end, deque.container.isFull, deque.container1.isFull, deque.isFull, deque.state)
	deque.RemoveLast()
	fmt.Println(deque.container.start, deque.container.end, deque.container1.start, deque.container1.end, deque.container.isFull, deque.container1.isFull, deque.isFull, deque.state)
	deque.AddFirst(1550.0)
	fmt.Println(deque.container.start, deque.container.end, deque.container1.start, deque.container1.end, deque.container.isFull, deque.container1.isFull, deque.isFull, deque.state)
	deque.RemoveLast()
	fmt.Println(deque.container.start, deque.container.end, deque.container1.start, deque.container1.end, deque.container.isFull, deque.container1.isFull, deque.isFull, deque.state)
	deque.AddFirst(1550.0)
	fmt.Println(deque.IsFull())
	fmt.Println(deque.container.start, deque.container.end, deque.container1.start, deque.container1.end, deque.container.isFull, deque.container1.isFull, deque.isFull, deque.state)

	deque.Set(deque.Size() - 1, 41, 269, 1490, 0)
	fmt.Println(deque.Get(deque.Size() - 1, 41, 269))
}
