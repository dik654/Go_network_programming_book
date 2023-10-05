// Golang의 context를 이용하여 타임아웃을 구현하는 코드
package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
	// 현재시간에서 5초 후를 dl이라는 이름으로 선언
	dl := time.Now().Add(5 * time.Second)
	// 컨텍스트로 타임아웃 설정
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	defer cancel()

	var d net.Dialer
	// Control 함수 따로 설정하는 방법
	d.Control = func(_, _ string, _ syscall.RawConn) error {
		time.Sleep(5*time.Second + time.Millisecond)
		return nil
	}

	// 컨텍스트를 갖고 TCP 리스너에게 통신 시도
	conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")
	// 통신 연결 성공 시
	if err == nil {
		// 연결을 끊고
		conn.Close()
		// 에러 리턴 및 종료
		t.Fatal("connection did not time out")
	}

	// 연결 실패 에러 type assertion
	nErr, ok := err.(net.Error)
	// type assertion 실패
	if !ok {
		t.Error(err)
	} else {
		// Listener의 에러가 시간 초과 에러가 아닐 때
		if !nErr.Timeout() {
			t.Errorf("error is not a timeout: %v", err)
		}
	}
	// context에러가 시간초과
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}
}
