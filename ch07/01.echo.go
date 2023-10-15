// 외부가 아닌 서비스간 통신을 위한
// 유닉스 도메인 소켓 "unix" 사용하는 코드ㅇ
package echo

import (
	"context"
	"net"
	"os"
)

func streamingEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	// addr 주소로 서버 생성
	// network에 tcp가 아닌 unix, unixpacket으로 유닉스 소켓 연결 가능
	s, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	// 고루틴 생성
	go func() {
		// 새로운 고루틴을 생성해서 컨텍스트 취소 시 리스너를 닫도록 대기시킴
		go func() {
			// 컨텍스트 취소 시
			<-ctx.Done()
			// 리스너 닫기
			_ = s.Close()
		}()

		for {
			// 리스너 연결 준비 완료가 되면
			conn, err := s.Accept()
			if err != nil {
				return
			}

			// 고루틴을 생성해서
			go func() {
				defer func() { _ = conn.Close() }()

				for {
					// 1KB 버퍼를 생성하여
					buf := make([]byte, 1024)
					// 서버로부터 데이터를 읽어 버퍼에 쓰고
					n, err := conn.Read(buf)
					if err != nil {
						return
					}
					// 버퍼에 썼던 데이터를 서버에 쓰기를 반복
					_, err = conn.Write(buf[:n])
					if err != nil {
						return
					}
				}
			}()
		}
	}()

	// 고루틴과 관계없이 즉시 서버의 주소를 리턴하고 종료
	return s.Addr(), nil
}

func datagramEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	// 인수로 들어온 네트워킹 타입에 맞게 리스너 생성
	s, err := net.ListenPacket(network, addr)
	if err != nil {
		return nil, err
	}

	// 고루틴 생성
	go func() {
		// 컨텍스트 취소용 고루틴 생성
		go func() {
			// 컨텍스트 취소 대기
			<-ctx.Done()
			// 리스너 종료
			_ = s.Close()
			// 유닉스 네트워크 타입이라면
			if network == "unixgram" {
				// 해당 경로의 파일, 디렉터리 삭제
				_ = os.Remove(addr)
			}
		}()

		// 1KB 버퍼 생성
		buf := make([]byte, 1024)
		for {
			// 리스너에 데이터가 들어오면 버퍼에 저장
			n, clientAddr, err := s.ReadFrom(buf)
			if err != nil {
				return
			}

			// 클라이언트 주소로 버퍼 내용 보내기
			_, err = s.WriteTo(buf[:n], clientAddr)
			if err != nil {
				return
			}
		}
	}()
	return s.LocalAddr(), nil
}
