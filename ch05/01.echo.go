// 송신자가 보낸 UDP 패킷을 echo해주는 UDP server
package echo

import (
	"context"
	"fmt"
	"net"
)

func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
	// UDP 리스너 생성
	// addr로 들어오는 패킷을 읽음
	s, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("binding to udp %s: %w", addr, err)
	}

	// 리스너 생성에 성공했다면 고루틴 생성
	go func() {
		// 추가 고루틴을 생성하여
		go func() {
			// 컨텍스트가 종료될 때까지 대기
			<-ctx.Done()
			// 컨텍스트가 종료됐다면 리스너 닫기
			_ = s.Close()
		}()

		// 1KB 버퍼 생성
		buf := make([]byte, 1024)

		for {
			// 리스너에서 데이터를 읽어서 버퍼에 저장
			// clientAddr는 데이터를 보낸 송신자
			n, clientAddr, err := s.ReadFrom(buf)
			if err != nil {
				return
			}

			// 읽은 데이터를 송신자(clientAddr)에게 쓰기
			_, err = s.WriteTo(buf[:n], clientAddr)
			if err != nil {
				return
			}
		}
	}()

	// 고루틴 생성과 관계없이 즉시 UDP 리스너의 주소 리턴
	return s.LocalAddr(), nil
}
