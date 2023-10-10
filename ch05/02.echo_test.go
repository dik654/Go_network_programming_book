// 클라이언트가 데이터를 보냈을 때
// 서버가 이를 읽고 클라이언트에게 동일한 데이터를 전송하는지 테스트
package echo

import (
	"bytes"
	"context"
	"net"
	"testing"
)

func TestEchoServerUDP(t *testing.T) {
	// 취소 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())
	// 01에서 만든 UDP 리스너 생성
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료시 취소 컨텍스트 실행
	defer cancel()

	// 클라이언트 UDP 리스너 생성
	// 클라이언트도 ListenPacket 사용 가능
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료시 클라이언트 연결 해제
	defer func() { _ = client.Close() }()

	// 보낼 ping이라는 문자의 byte데이터
	msg := []byte("ping")
	// 송신자의 주소를 가지고 UDP 서버에 데이터를 전송
	_, err = client.WriteTo(msg, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	// 1KB 버퍼 생성
	buf := make([]byte, 1024)
	// 클라이언트에게 들어온 데이터를 읽어서 버퍼에 저장
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// addr가 UDP 에코 리스너의 주소가 아니라면
	if addr.String() != serverAddr.String() {
		// 테스트 실패
		t.Fatalf("received reply from %q instead of %q", addr, serverAddr)
	}

	// 클라이언트가 보냈던 msg와 받은 데이터(버퍼)가 서로 다르다면
	if !bytes.Equal(msg, buf[:n]) {
		// 테스트 실패
		t.Errorf("expected reply %q; actual reply %q", msg, buf[:n])
	}
}
