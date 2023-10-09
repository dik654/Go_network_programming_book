// server => 서버로 ping이 들어오면 pong으로 변환
// proxyServer => 서버와 클라이언트 간 프록시 서버 연결용

package ch04

import (
	"io"
	"net"
	"sync"
	"testing"
)

// reader writer간 프록시
func proxy(from io.Reader, to io.Writer) error {
	fromWriter, fromIsWriter := from.(io.Writer)
	toReader, toIsReader := to.(io.Reader)
	// writer, reader가 맞다면
	if toIsReader && fromIsWriter {
		// 고루틴으로 fromWriter의 데이터 toReader로 복사
		go func() { _, _ = io.Copy(fromWriter, toReader) }()
	}
	// io.Writer의 데이터를 io.Reader로 복사
	_, err := io.Copy(to, from)

	return err
}

func TestProxy(t *testing.T) {
	// waitgroup 생성
	var wg sync.WaitGroup

	// 리스너 생성
	server, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// waitgroup 1개 추가
	wg.Add(1)

	// 고루틴 생성
	go func() {
		// 고루틴 종료 전 wg -1시키기
		defer wg.Done()

		for {
			// 리스너 연결 준비 완료
			conn, err := server.Accept()
			if err != nil {
				return
			}

			// 고루틴에서 고루틴 또 생성
			// 준비된 리스너 넘기기
			go func(c net.Conn) {
				// 이 고루틴이 종료되기 전 Accept 취소하기
				defer c.Close()

				for {
					// 1KB 버퍼 생성
					buf := make([]byte, 1024)
					// 리스너에서 값을 읽어서 buf에 저장
					n, err := c.Read(buf)
					if err != nil {
						// 읽다가 모든 데이터를 읽었으면
						if err != io.EOF {
							t.Error(buf)
						}
						// 이 고루틴 종료
						return
					}

					// 읽은 데이터를 string으로 변환해서
					switch msg := string(buf[:n]); msg {
					// 읽은 데이터가 "ping"이라면
					case "ping":
						// 리스너에 ping쓰기
						_, err = c.Write([]byte("pong"))
					default:
						// 아니라면 읽어온 그대로 쓰기
						_, err = c.Write(buf[:n])
					}
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}

						return
					}
				}
			}(conn)
		}
	}()

	// 프록시 서버 생성
	proxyServer, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// waitgroup 1개 추가
	wg.Add(1)

	go func() {
		// 고루틴 종료 전 wg -1시키기
		defer wg.Done()

		for {
			// 리스너 연결 준비 완료
			// 연결 요청하는 클라이언트를 순차적으로 연결해준다
			conn, err := proxyServer.Accept()
			if err != nil {
				return
			}

			// 고루틴에서 고루틴 또 생성
			// 리스너를 인수로 넘기기
			go func(from net.Conn) {
				// 이 고루틴 종료 시 Accept 취소하기
				defer from.Close()

				// 리스너에 연결 시도
				to, err := net.Dial("tcp", server.Addr().String())
				if err != nil {
					t.Error(err)
					return
				}
				// 이 고루틴 종료시 연결
				defer to.Close()

				// from 서버와 to 클라이언트 간
				// 프록시 연결 시도
				err = proxy(from, to)
				if err != nil && err != io.EOF {
					t.Error(err)
				}
			}(conn)
		}
	}()

	// 리스너에 연결 시도
	conn, err := net.Dial("tcp", proxyServer.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	msgs := []struct{ Message, Reply string }{
		{"ping", "pong"},
		{"pong", "pong"},
		{"echo", "echo"},
		{"ping", "pong"},
	}

	// i = index, m = message
	for i, m := range msgs {
		// 리스너에 message 쓰기
		_, err = conn.Write([]byte(m.Message))
		if err != nil {
			t.Fatal(err)
		}
		// 1KB 버퍼 생성
		buf := make([]byte, 1024)

		// 리스너에서 읽어서 buf에 쓰기
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		// 읽어온 데이터 string으로 변환
		actual := string(buf[:n])
		// 실제 메세지와 읽어온 데이터 로깅
		t.Logf("%q -> proxy -> %q", m.Message, actual)

		// 실제 메세지와 읽어온 데이터가 다르면
		if actual != m.Reply {
			// 에러 로깅
			t.Errorf("%d: expected reply: %q; actual: %q", i, m.Reply, actual)
		}
	}

	// 리스너 연결 해제
	_ = conn.Close()
	// 프록시서버 닫기
	_ = proxyServer.Close()
	// 서버 닫기
	_ = server.Close()

	// waitgroup이 0이 될 때까지 대기
	wg.Wait()
}

// go test -race -v proxy-test.go
