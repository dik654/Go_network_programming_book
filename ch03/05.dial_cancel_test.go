// 컨텍스트를 취소해서 Listener 연결을 취소하는 코드
package ch03

import (
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContextCancel(t *testing.T) {
	// 취소가 가능한 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())
	// 채널 생성
	sync := make(chan struct{})

	// 고루틴 선언
	go func() {
		// 고루틴 종료 시 Any역할을 하는 struct{}{}을 sync 채널에 전달
		defer func() { sync <- struct{}{} }()
		var d net.Dialer
		d.Control = func(_, _ string, _ syscall.RawConn) error {
			// 1초 대기 후
			time.Sleep(time.Second)
			// 종료
			return nil
		}

		// 생성한 컨텍스트로 listener 연결 시도
		conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:80")
		// 연결 실패 시
		if err != nil {
			// 에러 로깅 후 종료
			t.Log(err)
			return
		}
		// 연결 성공 시 연결을 끊고
		conn.Close()
		// 에러 리턴
		t.Error("connection did not time out")
	}()
	// 컨텍스트 취소하기
	cancel()
	// 채널에서 struct{}{} 데이터 꺼내기(따로 저장하진 않음)
	<-sync

	// 컨텍스트의 에러 사유가 cancel이 아니라면 에러 리턴
	if ctx.Err() != context.Cancel {
		t.Errorf("expect canceld context; actual: %q", ctx.Err())
	}
}
