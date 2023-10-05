// 패킷 송수신 없는동안
// 얼마까지 네트워크가 유지되는 시간제어
package ch03

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestDeadline(t *testing.T) {
	sync := make(chan struct{})
	// TCP 리스너 생성
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	//리스닝 고루틴
	go func() {
		//리스너와 연결하여 연결객체 리턴받기
		conn, err := listener.Accept()
		//연결 실패시 에러를 리턴하고 종료
		if err != nil {
			t.Log(err)
			return
		}
		//고루틴 종료시 리스너와의 연결 끊고 sync 채널 닫기
		defer func() {
			conn.Close()
			close(sync)
		}()
		//5초 이후를 데드라인으로 설정. 5초가 지날경우 err로 Timeout() 리턴
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}
		//1byte크기의 버퍼 생성
		buf := make([]byte, 1)
		//읽은 데이터를 buf에 저장(_는 읽은 바이트수)
		//Dialer가 데이터를 보낼 때까지 대기
		_, err = conn.Read(buf)
		//()는 type assertion으로 err가 net.Error interface를 구현하고 있는지 체크
		//nErr는 err의 실제 값
		nErr, ok := err.(net.Error)
		//Timeout에러가 아닌 경우 예외처리
		//ok가 false면 net.Error 인터페이스와 맞지않는 상태로 nErr는 비어있는 상태
		if !ok || !nErr.Timeout() {
			t.Errorf("expected timeout error; actual: %v", err)
		}
		//빈 구조체를 생성해서 sync 채널로 전송 (동기화, 고루틴 간 신호 전달용)
		sync <- struct{}{}

		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}
		//다음 데이터 올 때까지 대기
		_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
		}
	}()
	//리스너에게 tcp요청 보내서 연결 객체 리턴받기
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	//여기 도착 전까지 고루틴은 50번째 줄에서 대기
	//즉, 리스너와 연결되기 전까지 대기
	<-sync
	//연결된 tcp로 1이라는 데이터 보내기
	_, err = conn.Write([]byte("1"))
	if err != nil {
		t.Fatal(err)
	}
	//버퍼용 1바이트 슬라이스 생성
	buf := make([]byte, 1)
	//고루틴에서 데이터가 올 때까지 대기
	_, err = conn.Read(buf)
	//%v는 타입에 맞는 적절한 문자열 생성
	if err != io.EOF {
		t.Errorf("expected server termination; actual: %v", err)
	}
}
