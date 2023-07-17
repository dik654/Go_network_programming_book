package ch03

import (
	"net"
	"testing"
)

// 테스트를 위해 테스트 객체 인수를 받음
func TestListener(t testing.T) {
	// 127.0.0.1:남은포트번호의 tcp 상태로 listener 바인딩
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	// 연결이 제대로 되었는지 테스트
	if err != nil {
		t.Fatal(err)
	}
	// TestListner함수의 모든 처리가 끝나면 listener를 닫는다
	defer func() { _ = listener.Close() }()
	// 테스트 로그 콘솔에 print
	t.Logf("bound to %q", listener.Addr())
}
