package conpool

import (
	"sync"
	"testing"
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
	wg, _ := p[1].(*sync.WaitGroup)
	for i := 0; i < 500; i++ {
		c <- i
		<-c
	}
	wg.Done()
}

func BenchmarkConPool(b *testing.B) {
	pool := New(&ConPoolConfig{
		IdleWorkers: 10,
		MaxWorkers:  10,
	})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 10; j++ {
			wg.Add(1)
			c := make(chan int, 1)
			pool.Submit(senderMock, c, &wg)
		}
		wg.Wait()
		// time.Sleep(time.Millisecond * 20)
	}
}

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 10; j++ {
			wg.Add(1)
			c := make(chan int, 1)
			go senderMock(c, &wg)
		}
		wg.Wait()
		// time.Sleep(time.Millisecond * 20)
	}
}
