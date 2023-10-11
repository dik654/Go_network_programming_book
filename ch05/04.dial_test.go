// net.Dial을 사용하는 예제
package echo

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

func TestDialUDP(t *testing.T) {
	// 취소 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())
	// 에코서버 리스너 생성
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	defer cancel()
	// Dial을 이용하여 에코서버에 udp연결 클라
	client, err := net.Dial("udp", serverAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료시 연결 해제
	defer func() { _ = client.Close() }()

	// 끼어드는 연결 리스너 생성
	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// 클라이언트에 pardon me 보내기
	interrupt := []byte("pardon me")
	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	// 보낸 후 끼어드는 리스너 닫기
	_ = interloper.Close()

	// 클라이언트에 pardon me를 제대로 보냈는지 확인(클라이언트가 안 읽는다는 것을 확인하기 위해)
	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}

	// 클라이언트에서 에코서버로 ping보내기
	ping := []byte("ping")
	_, err = client.Write(ping)
	if err != nil {
		t.Fatal(err)
	}

	// 1KB 버퍼 생성
	buf := make([]byte, 1024)
	// 에코 서버에서 데이터 읽어 버퍼에 저장
	n, err = client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	// 클라이언트로 ping이 들어왔는지 확인
	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("expected reply %q; actual reply %q", ping, buf[:n])
	}

	// 1초 뒤 클라이언트 종료 설정
	err = client.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	// 클라이언트 데이터 읽기
	// 끼어드는 리스너에서 보낸 pardon me는 받지않았음을 알 수 있다
	_, err = client.Read(buf)
	if err == nil {
		t.Fatal("unexpected packet")
	}
}
