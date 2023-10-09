// 리스너에서 데이터 읽어오기
package ch04

import (
	"crypto/rand"
	"io"
	"net"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
	// 2^24 = 16,777,216 bytes
	// 16,777,216 / 1024 = 16,384KB
	// 16,384 / 1024 = 16MB
	// payload 버퍼 생성
	payload := make([]byte, 1<<24)
	// payload 버퍼에 랜덤한 값 넣기
	_, err := rand.Read(payload)
	// 랜덤 값 넣기에 실패하면 테스트 종료
	if err != nil {
		t.Fatal(err)
	}

	// 리스너 선언
	listener, err := net.Listen("tcp", "127.0.0.1:")
	// 선언 실패시 테스트 종료
	if err != nil {
		t.Fatal(err)
	}

	// 고루틴 생성
	go func() {
		// 리스너가 연결을 받을 준비 완료
		conn, err := listener.Accept()
		// 연결 준비 세팅 실패시 고루틴 종료
		if err != nil {
			t.Log(err)
			return
		}
		// 고루틴 종료시 리스너 연결 종료
		defer conn.Close()
		// payload에 들어있는 모든 데이터를 클라이언트로 쓰기
		_, err = conn.Write(payload)
		// 쓰기 실패시 에러 기록
		if err != nil {
			t.Error(err)
		}
	}()

	// 리스너에 연결 시도
	conn, err := net.Dial("tcp", listener.Addr().String())
	// 연결실패시 테스트 종료
	if err != nil {
		t.Fatal(err)
	}

	// 2^19 = 524,288 bytes
	// 524,288 / 1,024 = 512KB
	// 버퍼 생성
	buf := make([]byte, 1<<19)

	for {
		// 앞에서 Write한 payload
		// buf로 쓰기
		n, err := conn.Read(buf)
		// 쓰는 동안 에러가 날 경우
		if err != nil {
			// 모두 읽어서 나는 에러면
			if err != io.EOF {
				t.Error(err)
			}
			// 읽기 for문 종료
			break
		}
		t.Logf("read %d bytes", n)
	}
	// 모두 읽었으니 연결 종료
	conn.Close()
}
