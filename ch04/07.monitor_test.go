// 리스너 입출력을 모니터링
// Output:
// monitor: Test
// monitor: Test
package ch04

import (
	"io"
	"log"
	"net"
	"os"
)

// 모니터 타입 선언
type Monitor struct {
	*log.Logger
}

// 모니터에 Write 메서드 선언
func (m *Monitor) Write(p []byte) (int, error) {
	// 쓰려는 byte의 길이
	// m.Ouput 메서드는 *log.Logger 타입의 내장 메서드
	// 2는 함수의 호출 깊이, 로그 메시지가 발생한 위치를 정확히 추적하기 위해 사용
	// string(p)는 인수로 들어온 byte를 string으로 변환하여, 출력될 로그 메시지 문자열
	return len(p), m.Output(2, string(p))
}

func ExampleMonitor() {
	// 모니터 인스턴스 생성
	// log.New(
	//			로그 메시지가 출력될 io.Writer,
	//			각 로그 메시지 앞에 추가될 접두사
	//			log.Ldate, log.Ltime, log.Lshortfile 등 플래그
	// )
	monitor := &Monitor{Logger: log.New(os.Stdout, "monitor: ", 0)}

	// 리스너 생성
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		monitor.Fatal(err)
	}

	// done 채널 생성
	done := make(chan struct{})

	// 고루틴 생성
	go func() {
		// 고루틴 종료시 done 채널 닫기
		defer close(done)

		// 리스너 열기
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		// 종료시 리스너 닫기
		defer conn.Close()

		// 1KB 버퍼 생성
		b := make([]byte, 1024)
		// 읽는 동시에 쓰는 reader 생성, conn을 읽으면 monitor에 써짐
		r := io.TeeReader(conn, monitor)

		// 리스너 연결로부터 읽은 데이터를 버퍼, monitor에 쓴다
		// 첫 번째 monitor: Test
		n, err := r.Read(b)
		// 끝에 도달하면 if문 패스
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}

		// 여러 Writer를 결합해 하나처럼 동작
		// conn와 monitor가 동시에 쓰기 작업이 됨
		// 62번 줄에서 conn에 보내기만하면 monitor에 써지는데 이렇게 했는지 모르겠음
		w := io.MultiWriter(conn, monitor)

		// 1KB 버퍼에서 읽어서 conn과 monitor에 쓰기
		// 두 번째 monitor: Test
		_, err = w.Write(b[:n])
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}
	}()

	// 리스너에 연결 시도
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		monitor.Fatal(err)
	}

	// 리스너에 "Test" 보내기
	_, err = conn.Write([]byte("Test\n"))
	if err != nil {
		monitor.Fatal(err)
	}

	// 리스너와 연결 해제
	_ = conn.Close()
	// done 채널 닫기
	<-done
}
