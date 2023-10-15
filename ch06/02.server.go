// TFTP 서버 코드
package tftp

import (
	"bytes"
	"errors"
	"log"
	"net"
	"time"
)

// 서버에서 연결 관리를 위해 필요한 데이터 구조체
type Server struct {
	Payload []byte
	// 재시도 횟수
	Retries uint8
	// 연결 종료 시간
	Timeout time.Duration
}

func (s Server) ListenAndServe(addr string) error {
	// udp 리스너 생성
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	// 리스너 종료
	defer func() { _ = conn.Close() }()

	// 리스너 주소 콘솔에 쓰기
	log.Printf("Listening on %s ...\n", conn.LocalAddr())

	return s.Serve(conn)
}

func (s *Server) Serve(conn net.PacketConn) error {
	// conn net.PacketConn는 udp 연결
	// conn == nil은 연결이 없는 경우 에러
	if conn == nil {
		return errors.New("nil connection")
	}

	// 서버에 payload가 없는 경우에도 에러
	if s.Payload == nil {
		return errors.New("payload is required")
	}

	// 남은 재시도 횟수가 0이면 10으로 초기화
	if s.Retries == 0 {
		s.Retries = 10
	}

	// Timeout이 0으로 만료됐으면 6초로 초기화
	if s.Timeout == 0 {
		s.Timeout = 6 * time.Second
	}

	var rrq ReadReq

	for {
		// 데이터그램 크기만큼 버퍼 생성
		buf := make([]byte, DatagramSize)

		// conn에서 데이터 읽어서 buf에 저장
		// addr에 데이터 송신자 address 저장
		_, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return err
		}

		// rrq에 버퍼에 적힌 패킷내용 옮기기
		// opcode, 파일명, mode
		err = rrq.UnmarshalBinary(buf)
		if err != nil {
			log.Printf("[%s] bad request: %v", addr, err)
			continue
		}

		// 패킷의 종류에 맞춰 동작하도록 handle메서드 실행
		go s.handle(addr.String(), rrq)
	}
}

func (s Server) handle(clientAddr string, rrq ReadReq) {
	log.Printf("[%s] request file: %s", clientAddr, rrq.Filename)

	// 클라 주소로 udp 연결
	conn, err := net.Dial("udp", clientAddr)
	if err != nil {
		log.Printf("[%s] dial: %v", clientAddr, err)
	}
	// 함수 종료시 udp 연결 끊기
	defer func() { _ = conn.Close() }()

	var (
		ackPkt  Ack
		errPkt  Err
		dataPkt = Data{Payload: bytes.NewReader(s.Payload)}
		buf     = make([]byte, DatagramSize)
	)

NEXTPACKET:
	// continue에서 반복하다가 보낸 데이터의 크기가 516보다 작아지면 for문 종료
	for n := DatagramSize; n == DatagramSize; {
		data, err := dataPkt.MarshalBinary()
		if err != nil {
			log.Printf("[%s] preparing data packet: %v", clientAddr, err)
		}

	RETRY:
		// 재연결 시도 횟수
		for i := s.Retries; i > 0; i-- {
			// 클라이언트에 data 보내기
			// Datagram 크기로 초기화되었던 n을 보낸 data 크기로 업데이트
			n, err = conn.Write(data)
			if err != nil {
				log.Printf("[%s] write: %v", clientAddr, err)
				return
			}

			// 연결에 Timeout 변수만큼 데드라인 설정
			_ = conn.SetReadDeadline(time.Now().Add(s.Timeout))

			// 클라이언트에 들어온 데이터 버퍼에 복사
			_, err = conn.Read(buf)
			if err != nil {
				if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
					continue RETRY
				}

				log.Printf("[%s] waiting for ACK: %v", clientAddr, err)
				return
			}

			switch {
			// ackPkt에 블록 번호 저장
			case ackPkt.UnmarshalBinary(buf) == nil:
				// 읽은 블록번호가 데이터패킷에 들어있는 데이터 번호랑 같다면
				// NEXTPACKET 계속 실행
				if uint16(ackPkt) == dataPkt.Block {
					continue NEXTPACKET
				}
			// 에러코드 언마샬링에 성공한 경우
			// errPkt.Message에 메세지 저장
			// errPkt.Error에 에러코드 저장
			case errPkt.UnmarshalBinary(buf) == nil:
				log.Printf("[%s] received error: %v", clientAddr, errPkt.Message)
				return
			default:
				// 언마샬링 모두 실패시 잘못된 패킷이라고 리턴
				log.Printf("[%s] bad packet", clientAddr)
			}
		}

		log.Printf("[%s] exhausted retries", clientAddr)
		return
	}

	log.Printf("[%s] sent %d blocks", clientAddr, dataPkt.Block)
}
