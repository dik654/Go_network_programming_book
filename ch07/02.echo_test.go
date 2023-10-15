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

func TestEchoServerUnix(t *testing.T) {
	// ioutil.TempDir("", "echo_unix") 대신 사용
	dir, err := os.MkdirTemp("", "echo_unix")
	if err != nil {
		t.Fatal(err)
	}
	// 서버 종료시
	defer func() {
		// 디렉터리 삭제
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	// 취소 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())
	// os.Getpid() = 현재 실행 중인 프로세스의 ID를 가져오기
	// 유닉스 소켓 경로 생성 = /디렉터리명/23490.sock
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	// 유닉스 소켓 에코 서버 생성
	rAddr, err := streamingEchoServer(ctx, "unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료시 컨텍스트 취소
	defer cancel()

	// 소켓 파일의 권한을 모든 유저로 변경
	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	// 에코 서버에 연결 시도
	conn, err := net.Dial("unix", rAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료시 연결 닫기
	defer func() { _ = conn.Close() }()

	msg := []byte("ping")
	// 3번 반복
	for i := 0; i < 3; i++ {
		// 에코 서버에 ping 보내기
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	// 1KB 버퍼 생성
	buf := make([]byte, 1024)
	// 에코 서버에서 받은 데이터 버퍼에 쓰기
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := bytes.Repeat(msg, 3)
	// 에코에 ping이 세 번 들어오지 않았다면
	if !bytes.Equal(expected, buf[:n]) {
		// 테스트 실패
		t.Fatalf("expected reply %q; actual reply %q", expected, buf[:n])
	}
}
