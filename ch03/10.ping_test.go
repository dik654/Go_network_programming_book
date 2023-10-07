// 수신 데이터 데드라인 늦추기
package ch03

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestPingerAdvanceDeadline(t *testing.T) {
	// done 채널 생성
	done := make(chan struct{})
	// 리스너 생성
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// 시작시간 변수에 저장
	begin := time.Now()
	// 고루틴 시작
	go func() {
		// 고루틴 종료시 done 채널 닫기
		defer func() { close(done) }()
		// 리스너가 연결을 받을 준비
		conn, err := listener.Accept()
		// 연결 실패시 로깅 후 고루틴 종료
		if err != nil {
			t.Log(err)
			return
		}
		// 취소 컨텍스트 생성
		ctx, cancel := context.WithCancel(context.Background())

		defer func() {
			// 고루틴 종료시 컨텍스트 취소 및
			cancel()
			// 리스너 연결 끊기
			conn.Close()
		}()

		// 09에서 사용한 resetTimer와 동일한 방식
		resetTimer := make(chan time.Duration, 1)
		resetTimer <- time.Second
		go Pinger(ctx, conn, resetTimer)

		// 연결 제한시간 5초로 설정
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		// 제한시간 설정 중 에러가 있으면 함수 종료
		if err != nil {
			t.Error(err)
			return
		}

		// 버퍼 생성
		buf := make([]byte, 1024)
		for {
			// 리스너 연결된 내용을 버퍼에 저장
			n, err := conn.Read(buf)
			// 버퍼에 저장하는 도중 오류가 생기면 에러 리턴후 종료
			if err != nil {
				return
			}
			t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])

			// resetTimer채널에 0값 보내기
			// 핑 간격을 defaultPingInterval로 재설정하도록 Pinger 함수에 지시
			resetTimer <- 0
			// 연결 제한시간을 다시 5초로 설정
			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()

	// 리스너에 연결 시도
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	// 메인 함수 종료 후 리스너 연결 끊기
	defer conn.Close()

	// 버퍼 생성
	buf := make([]byte, 1024)
	// 핑 4개 읽기
	for i := 0; i < 4; i++ {
		// 리스너 연결에서 받은 데이터 버퍼에 저장
		n, err := conn.Read(buf)
		// 버퍼 저장 중 에러가 생기면 테스트 종료
		if err != nil {
			t.Fatal(err)
		}
		// Since는 begin으로부터 얼마나 지났는지
		// Truncate 초 단위로 자르기
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}

	// 리스너에 PONG 보내기
	_, err = conn.Write([]byte("PONG!!!"))
	// 실패시 테스트 종료
	if err != nil {
		t.Fatal(err)
	}
	// ping 4개 더 읽기
	for i := 0; i < 4; i++ {
		n, err := conn.Read(buf)
		// 모두 읽었으면 for문 탈출
		if err != nil {
			// 모두 읽었다는 에러가 아니라면 테스트 종료
			if err != io.EOF {
				t.Fatal(err)
			}
			break
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	// done 채널 닫기
	<-done
	end := time.Since(begin).Truncate(time.Second)
	t.Logf("[%s] done", end)
	if end != 9*time.Second {
		t.Fatalf("expected EOF at 9seconds; actual %s", end)
	}
}
