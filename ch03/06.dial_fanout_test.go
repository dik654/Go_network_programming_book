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
	// 10초 제한시간 컨텍스트 생성
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(10*time.Second),
	)

	// 리스너 생성
	listener, err := net.Listen("tcp", "127.0.0.1:")
	// 생성 실패시 에러 리턴 및 종료
	if err != nil {
		t.Fatal(err)
	}

	// 테스트 함수 종료시 리스너 종료
	defer listener.Close()

	// 고루틴 생성
	go func() {
		// 리스너 연결
		conn, err := listener.Accept()
		// 연결 성공시
		if err == nil {
			// 연결 끊기
			conn.Close()
		}
	}()

	// 고루틴용 함수 선언
	dial := func(ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup) {
		// 채널로 struct{}를 보내 고루틴이 종료됐음을 알림
		defer wg.Done()

		// Dialer 선언
		var d net.Dialer
		// 앞에서 선언한 컨텍스트를 이용하여 Listener에 연결
		c, err := d.DialContext(ctx, "tcp", address)
		// 연결 실패시 함수 종료
		if err != nil {
			return
		}
		// 연결 성공시 통신 끊기
		c.Close()

		// case들 중 실행 가능한 case를 랜덤으로 실행
		// 아래 case들 중 하나가 실행되기 전까지 여기서 대기
		select {
		// 컨텍스트 취소가 발생하면 실행
		case <-ctx.Done():
		// id를 response 채널에 넘기기
		// case response <- id:는 "response 채널이 열려있을 때" 실행되어 id 값을 전송하고 로직이 실행되는 case
		// case response := <-id:는 "id 값이 response 채널에 전송되었을 때"  로직이 실행되는 case
		case response <- id:
		}
	}

	res := make(chan int)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		// 고루틴이 생성될 때마다 wg 1개씩 추가
		wg.Add(1)
		// 앞에서 선언한 dial함수 실행하여 Listener 연결 여러번 시도
		go dial(ctx, listener.Addr().String(), res, i+1, &wg)
	}

	// 고루틴 함수에서 response 채널에 넘긴 id 값 response 변수에 저장
	// 값이 들어올 때까지 여기서 멈춰서 대기
	response := <-res
	// 10초 제한시간이 있지만 10초가 되기 전에 컨텍스트 cancel 실행
	cancel()
	// 고루틴들이 모두 종료될 때까지 대기
	wg.Wait()
	// res 채널 닫기
	close(res)

	// ex)
	// ch := make(chan int)
	// close(ch)
	// v, ok := <-ch 에서 v는 0, ok는 false값을 가짐

	// cancel 함수에 의한 컨텍스트 에러가 아니면 에러 리턴
	if ctx.Err() != context.Canceled {
		t.Errorf("expected canceled context; acutal; %s", ctx.Err())
	}

	// response 변수 로깅
	t.Logf("dialer %d retrieved the resource", response)
}
