package conpool

import (
	"testing"
	"time"
)

// func TestConPool(t *testing.T) {
// 	err := Submit(func(params ...interface{}) {
// 		fmt.Println(params)
// 		time.Sleep(time.Second)
// 	}, 1, "string", false)

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	time.Sleep(time.Second)
// }

func senderMock(p ...interface{}) {
	c, _ := p[0].(chan int)
	for i := 0; i < 500; i++ {
		c <- i
		<-c
	}
}

func BenchmarkConPool(b *testing.B) {
	pool := New(&ConPoolConfig{
		IdleWorkers: 10,
		MaxWorkers:  10,
	})

	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			c := make(chan int, 1)
			pool.Submit(senderMock, c)
		}
		time.Sleep(time.Millisecond * 20)
	}
}

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			c := make(chan int, 1)
			go senderMock(c)
		}
		time.Sleep(time.Millisecond * 20)
	}
}
