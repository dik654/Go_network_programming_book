// Listener(TCP서버)와 연결하는 과정
package ch03

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// 임의의 포트로 바인딩하여 리스너 생성
	listener, err := net.Listen("tcp", "127.0.0.1:")
	// 리스너 생성 성공 여부 테스트
	if err != nil {
		t.Fatal(err)
	}
	// 종료 신호용 채널 생성
	done := make(chan struct{})
	go func() {
		// 고루틴 종료시 done에 종료 신호
		defer func() { done <- struct{}{} }()

		for {
			// 리스너와 고루틴간의 연결 수립
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}
			// 연결 수립된 객체를 넘겨서 고루틴 생성
			go func(c net.Conn) {
				defer func() {
					// 고루틴에서 작업이 완료되면 연결을 닫고, done에 종료 신호
					c.Close()
					done <- struct{}{}
				}()
				// 1024bytes 버퍼용 슬라이스 생성
				buf := make([]byte, 1024)
				for {
					// 들어온 요청을 버퍼에 담아서 읽기
					n, err := c.Read(buf)
					//읽는 도중 에러가 발생했을 때
					if err != nil {
						// 파일의 끝이 아니라면 에러 내용을 리턴
						if err != io.EOF {
							t.Error(err)
						}
						// 파일의 끝이라면 고루틴 종료
						return
					}

					t.Logf("received %q", buf[n])
				}
			}(conn)

		}
	}()
	// tcp 연결객체 리턴
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	// 연결 닫기
	conn.Close()
	// 30줄 고루틴에서 done 채널로 전송(defer)할 때까지 대기
	<-done
	// 리스너 닫기
	listener.Close()
	// 18줄 고루틴에서 done 채널로 전송(defer)할 때까지 대기
	<-done
}
