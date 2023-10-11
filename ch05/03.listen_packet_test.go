// 여러 ListenPacket끼리 통신이 가능하다는 것을 보여주는 예제
package echo

import (
	"bytes"
	"context"
	"net"
	"testing"
)

func TestListenPacketUDP(t *testing.T) {
	// 취소 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())
	// 에코 서버 리스너 생성
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료시 컨텍스트 취소하여 리스너 닫기
	defer cancel()

	// 클라이언트 리스너 생성
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료시 클라이언트 리스너 닫기
	defer func() { _ = client.Close() }()

	// 에코 서버와 클라이언트 사이에 끼어드는 리스너 생성
	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// 끼어드는 데이터 pardon me
	interrupt := []byte("pardon me")
	// 끼어드는 리스너로에서 클라이언트로 pardon me 전송하기
	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	// 이후 끼어드는 리스너 닫기
	_ = interloper.Close()

	// 보낸 bytes수랑 pardon me랑 길이가 다르면 테스트 실패
	if l := len(interrupt); l != n {
		t.Fatalf("wrote %d bytes of %d", n, l)
	}

	// 클라이언트에서 에코 서버로 ping 전송
	ping := []byte("ping")
	_, err = client.WriteTo(ping, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	// 1KB 버퍼 생성
	buf := make([]byte, 1024)
	// 클라이언트로 들어온 데이터 버퍼에 저장
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// 클라에 들어온 데이터가 pardon me가 아니면
	if !bytes.Equal(interrupt, buf[:n]) {
		// 에러
		t.Errorf("expected reply %q; actual reply %q", interrupt, buf[:n])
	}

	// 클라에 데이터를 보낸 주소가 끼어드는 리스너가 아니라면
	if addr.String() != interloper.LocalAddr().String() {
		// 에러
		t.Errorf("expected message from %q; actual sender is %q", interloper.LocalAddr(), addr)
	}

	// 클라이언트 데이터 읽어서 버퍼에 저장
	n, addr, err = client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// 데이터가 ping이 아니라면
	if !bytes.Equal(ping, buf[:n]) {
		// 에러
		t.Errorf("expected reply %q; actual reply %q", ping, buf[:n])
	}

	// 서버에서 온 데이터가 아니라면
	if addr.String() != serverAddr.String() {
		에러
		t.Errorf("expected message from %q; actual sender is %q", serverAddr, addr)
	}
}
