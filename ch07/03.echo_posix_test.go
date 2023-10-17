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

func TestEchoServerUnixDatagram(t *testing.T) {
	// 시스템 임시 디렉터리에 echo_unixgram 디렉터리 생성
	dir, err := os.MkdirTemp("", "echo_unixgram")
	if err != nil {
		t.Fatal(err)
	}

	// 테스트 종료시 생성한 echo_unixgram 디렉터리 삭제
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	// 취소 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())
	// echo_unixgram 디렉터리에 pid 이름을 갖는 서버용 유닉스 소켓파일명
	sSocket := filepath.Join(dir, fmt.Sprintf("s%d.sock", os.Getpid()))
	// sSocket을 이용해서 unixgram 네트워크 서버 생성
	// 생성 시 소켓 파일이 생성됨
	serverAddr, err := datagramEchoServer(ctx, "unixgram", sSocket)
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료시 컨텍스트 취소
	defer cancel()

	// 소켓파일의 권한 변경
	err = os.Chmod(sSocket, os.ModeSocket|0622)
	if err != nil {
		t.Fatal(err)
	}

	// 클라이언트용 유닉스 소켓파일명
	cSocket := filepath.Join(dir, fmt.Sprintf("c%d.sock", os.Getpid()))
	// 클라이언트 소켓 리스너 생성
	// 이 때 소켓 파일이 생성됨
	client, err := net.ListenPacket("unixgram", cSocket)
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료 시 리스너 닫기
	defer func() { _ = client.Close() }()

	// 소켓파일의 권한 변경
	err = os.Chmod(cSocket, os.ModeSocket|6222)
	if err != nil {
		t.Fatal(err)
	}

	// 클라이언트 주소에서 서버로 ping 3개 보내기
	msg := []byte("ping")
	for i := 0; i < 3; i++ {
		_, err = client.WriteTo(msg, serverAddr)
		if err != nil {
			t.Fatal(err)
		}
	}

	// 1KB 버퍼 생성
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		// 클라이언트에 들어온 데이터 버퍼에 쓰기
		n, addr, err := client.ReadFrom(buf)
		if err != nil {
			t.Fatal()
		}

		// 클라이언트에 데이터를 보낸 송신자가 서버가 아니면
		if addr.String() != serverAddr.String() {
			// 테스트 실패
			t.Fatalf("received reply from %q instead of %q", addr, serverAddr)
		}

		// 메세지의 길이가 다르면
		if !bytes.Equal(msg, buf[:n]) {
			// 테스트 실패
			t.Fatalf("expected reply %q; actual reply %q", msg, buf[:n])
		}
	}
}
