// 한 컨텍스트를 여러 함수에게 전달
// 해당 컨텍스트를 cancel하면 여러 함수의 Dial을 한꺼번에 종료할 수 있다
package ch03

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextCancelFanOut(t *testing.T) {
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(10*time.Second),
	)

	listener, err := net.Listener("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		conn, err := listner.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	dial := func(ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup) {
		defer wg.Done()

		var d net.Dialer
		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}
		c.Close()

		select {
		case <-ctx.Done():
		case response <- id:
		}

		res := make(chan int)
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go dial(ctx, listner.Addr().String(), res, i+1, &wg)
		}

		response := <-res
		cancel()
		wg.Wait()
		close(res)

		if ctx.Err() != context.Canceled {
			t.Errorf("expected canceled context; acutal; %s", ctx.Err())
		}

		t.Logf("dialer %d retrieved the resource", response)
	}
}
