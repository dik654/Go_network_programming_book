package echo

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestEchoServerUnixPacket(t *testing.T) {
	dir, err := os.MkdirTemp("", "echo_unixpacket")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	// unixpacket은 각 패킷이 독립적
	// unix는 스트림으로 동작
	// 둘 다 순서, 신뢰성 보장O
	rAddr, err := streamingEchoServer(ctx, "unixpacket", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	// 소켓파일 권한 변경
	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	// 에코서버에 연결 시도
	conn, err := net.Dial("unixpacket", rAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료시 연결 끊기
	defer func() { _ = conn.Close() }()

	// 에코 서버에 ping 3번 보내기
	msg := []byte("ping")
	for i := 0; i < 3; i++ {
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	// 1KB 버퍼 생성
	buf := make([]byte, 1024)
	// 에코잉된 ping 3번 읽기
	for i := 0; i < 3; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(msg, buf[:n]) {
			t.Errorf("expected reply %q; actual reply %q", msg, buf[:n])
		}
	}

	// ping 3개 더 보내기
	for i := 0; i < 3; i++ { // write 3 more "ping" messages
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	// 2bytes 버퍼 생성
	buf = make([]byte, 2)
	for i := 0; i < 3; i++ {
		// 2bytes 읽기
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		// 실제로 읽은 데이터가 "pi"에 해당하는지 확인
		if !bytes.Equal(msg[:2], buf[:n]) {
			t.Errorf("expected reply %q; actual reply %q", msg[:2],
				buf[:n])
		}
	}
}

func BenchmarkEchoServerUnixPacket(b *testing.B) {
	dir, err := os.MkdirTemp("", "echo_unixpacket_bench")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			b.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	rAddr, err := streamingEchoServer(ctx, "unixpacket", socket)
	if err != nil {
		b.Fatal(err)
	}
	defer cancel()

	conn, err := net.Dial("unixpacket", rAddr.String())
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	msg := []byte("ping")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = conn.Write(msg)
		if err != nil {
			b.Fatal(err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			b.Fatal(err)
		}

		if !bytes.Equal(msg, buf[:n]) {
			b.Fatalf("expected reply %q; actual reply %q", msg, buf[:n])
		}
	}
}
