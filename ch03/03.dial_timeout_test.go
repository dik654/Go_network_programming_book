// Listener 연결이 일정시간되지않으면 실패시키는(timeout) 코드
package ch03

import (
	"net"
	"syscall"
	"testing"
	"time"
)

func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// DialTimeout자체가 Dialer를 가지고 있지 않아 net.Dialer 인터페이스로 구현체 선언
	d := net.Dialer{
		// 타임아웃시 에러 반환을 위해 Control overriding
		Control: func(_, addr string, _ syscall.RawConn) error {
			// DNS 에러 반환
			return &net.DNSError{
				Err:         "connection timeout",
				Name:        addr,
				Server:      "127.0.0.1",
				IsTimeout:   true,
				IsTemporary: true,
			}
		},
		// 얼마나 기다릴지 설정
		Timeout: timeout,
	}
	// dial_test에서 사용한 Dial 함수 리턴
	return d.Dial(network, address)
}

func TestDialTimeout(t *testing.T) {
	// TCP로 "10.0.0.1:http"에 5초 타임아웃을 가지고 접속시도
	c, err := DialTimeout("tcp", "10.0.0.1:http", 5*time.Second)

	// 정상적으로 연결됐다면
	if err == nil {
		// 연결 끊기
		c.Close()
		// 에러 리턴
		t.Fatal("connection did not time out")
	}

	// testing에 맞는 에러타입으로 type assertion
	nErr, ok := err.(net.Error)
	// type assertion 실패시
	if !ok {
		t.Fatal(err)
	}

	// 타임아웃에러가 아니라면 에러 리턴
	if !nErr.Timeout() {
		t.Fatal("error is not a timeout")
	}

	// 결과적으로 타임아웃 에러일 경우에만 통과
}
