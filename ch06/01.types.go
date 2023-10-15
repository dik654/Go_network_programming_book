package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	// 가능한 최대 데이터그램 크기
	DatagramSize = 516
	// 4bytes 헤더를 제외한 데이터그램 크기
	BlockSize = DatagramSize - 4
)

// TFTP의 첫 2bytes = opcode
type OpCode uint16

// opcode 선언
const (
	// Read ReQuest
	OpRRQ OpCode = iota + 1
	// 읽기 전용 서버를 구현할 것이므로
	// Write ReQuest(WRQ) 지원 x
	_
	OpData
	OpAck
	OpErr
)

type ErrCode uint16

// OpErr
// 에러코드 선언
const (
	ErrUnknown ErrCode = iota
	ErrNotFound
	ErrAccessViolation
	ErrDiskFull
	ErrIllegalOp
	ErrUnknownID
	ErrFileExists
	ErrNoUser
)

// 읽기 요청 정의
type ReadReq struct {
	Filename string
	// netascii mode - 파일을 line ending format에 맞춰 변환해줘야함
	// octet mode - 파일을 바이너리 그대로 보냄
	Mode string
}

// 핸드셰이크
func (q ReadReq) MarshalBinary() ([]byte, error) {
	// 기본적으로 octet모드 설정 준비
	mode := "octet"
	// 요청을 읽었을 때 mode가 정해져 있었다면
	if q.Mode != "" {
		// 해당 모드로 mode변수 변경
		mode = q.Mode
	}

	// opcode 2bytes
	// 파일명
	// 0 1bytes
	// mode 정보
	// 0 1bytes
	cap := 2 + len(q.Filename) + 1 + len(q.Mode) + 1

	// 버퍼 생성
	b := new(bytes.Buffer)
	// 버퍼 용량 늘리기
	b.Grow(cap)

	// 차례대로 버퍼에 써서
	// 읽기 요청 패킷 구조에 맞게 데이터 쌓기

	// 먼저 버퍼에 RRQ opcode쓰기
	err := binary.Write(b, binary.BigEndian, OpRRQ)
	if err != nil {
		return nil, err
	}

	// 위에서 쓴 RRQ opcode를 q.Filename에 쓰기
	_, err = b.WriteString(q.Filename)
	if err != nil {
		return nil, err
	}

	// 1byte 0쓰기
	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}

	// mode 쓰기
	_, err = b.WriteString(mode)
	if err != nil {
		return nil, err
	}

	// 1byte 0쓰기
	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}

	// 작성한 요청 버퍼 리턴
	return b.Bytes(), nil
}

// 위에서 만든 패킷을 읽을 수 있는 형태로 다시 해석
func (q *ReadReq) UnmarshalBinary(p []byte) error {
	// p내용을 갖고, p의 길이를 갖는 버퍼 생성
	r := bytes.NewBuffer(p)

	// uint16 = 2bytes
	var code OpCode

	// 빅 엔디언으로 code에 opcode(2bytes) 저장
	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return nil
	}

	// code에 OpRRQ가 써지지 않았다면
	if code != OpRRQ {
		return errors.New("invalid RRQ")
	}

	// 0을 만날 때까지 r에서 데이터 읽기, 즉 파일명 읽기
	q.Filename, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	// 파일명에 0까지 붙어있으므로 이를 제거
	q.Filename = strings.TrimRight(q.Filename, "\x00")
	if len(q.Filename) == 0 {
		return errors.New("invalid RRQ")
	}

	// 다음 0을 만날 때까지 r에서 데이터 읽어, mode 읽기
	q.Mode, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	// mode명에 0이 붙어있으므로 0을 제거
	q.Mode = strings.TrimRight(q.Mode, "\x00")
	if len(q.Mode) == 0 {
		return errors.New("invalid RRQ")
	}

	// 받은 mode 문자열을 소문자로 변경
	actual := strings.ToLower(q.Mode)
	if actual != "octet" {
		return errors.New("only binary transfers supported")
	}

	return nil
}

type Data struct {
	Block   uint16
	Payload io.Reader
}

// 실제 데이터 교환
func (d *Data) MarshalBinary() ([]byte, error) {
	// 버퍼 생성
	b := new(bytes.Buffer)
	// 데이터그램만큼 버퍼 크기 증가
	b.Grow(DatagramSize)

	// 블록 번호 증가
	d.Block++

	// 버퍼에 Opdata 쓰기
	err := binary.Write(b, binary.BigEndian, OpData)
	if err != nil {
		return nil, err
	}

	// 버퍼에 블록 높이 쓰기
	err = binary.Write(b, binary.BigEndian, d.Block)
	if err != nil {
		return nil, err
	}

	// b에서 d.Payload로 BlockSize bytes만큼 복사
	_, err = io.CopyN(b, d.Payload, BlockSize)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return b.Bytes(), nil
}

func (d *Data) UnmarshalBinary(p []byte) error {
	// 받은 bytes 데이터 길이가 4보다 작거나
	// Datagram 제한 길이보다 크다면 에러
	if l := len(p); l < 4 || l > DatagramSize {
		return errors.New("invalid DATA")
	}

	var opcode OpCode

	// 첫 2bytes 읽어서 opcode 변수에 저장
	err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &opcode)
	if err != nil || opcode != OpData {
		return errors.New("invalid DATA")
	}

	// 블록 번호 읽기
	err = binary.Read(bytes.NewReader(p[2:4]), binary.BigEndian, &d.Block)

	// 4bytes부터 payload받아오기
	d.Payload = bytes.NewBuffer(p[4:])

	return nil
}

// 블록 번호
type Ack uint16

// 정상적으로 받았다는 확인 데이터(수신 확인 패킷) 마샬링
func (a Ack) MarshalBinary() ([]byte, error) {
	// opcode + 블록 번호
	cap := 2 + 2

	// 버퍼 생성
	b := new(bytes.Buffer)
	// 4bytes만큼 버퍼 크기 증가
	b.Grow(cap)

	// 버퍼에 OpAck opcode 넣기
	err := binary.Write(b, binary.BigEndian, OpAck)
	if err != nil {
		return nil, err
	}

	// 쓰면 쓴 만큼 버퍼에서 사라지므로 블록 번호 쓰기
	err = binary.Write(b, binary.BigEndian, a)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (a *Ack) UnmarshalBinary(p []byte) error {
	var code OpCode
	// 변수 r에 수신 확인 패킷 p 저장
	r := bytes.NewReader(p)

	// code 변수에 opcode 저장
	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return err
	}

	// opcode가 OpAck가 아니라면 에러
	if code != OpAck {
		return errors.New("invalid ACK")
	}

	// 블록 번호 읽어서 a에 저장
	return binary.Read(r, binary.BigEndian, a)
}

type Err struct {
	Error   ErrCode
	Message string
}

// 에러 처리용 패킷 마샬링
func (e Err) MarshalBinary() ([]byte, error) {
	// opcode + 에러코드 + 메세지 바이트, 구분용 0
	cap := 2 + 2 + len(e.Message) + 1
	b := new(bytes.Buffer)
	b.Grow(cap)

	// 버퍼에 OpErr 쓰기
	err := binary.Write(b, binary.BigEndian, OpErr)
	if err != nil {
		return nil, err
	}

	// 버퍼에 실제 에러코드 쓰기
	err = binary.Write(b, binary.BigEndian, e.Error)
	if err != nil {
		return nil, err
	}

	// 메세지 바이트 쓰기
	_, err = b.WriteString(e.Message)
	if err != nil {
		return nil, err
	}

	// 구분용 0 쓰기
	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}

	// 마샬링한 에러 패킷 리턴
	return b.Bytes(), nil
}

func (e *Err) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)

	var code OpCode

	// opcode code에 저장
	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return nil
	}

	// opcode가 OpErr가 아니면 에러
	if code != OpErr {
		return errors.New("invalid ERROR")
	}

	// 에러코드 e.Error에 저장
	err = binary.Read(r, binary.BigEndian, &e.Error)
	if err != nil {
		return err
	}

	// 0이 나올 때까지 읽고
	e.Message, err = r.ReadString(0)
	// 오른쪽 0 지우기
	e.Message = strings.TrimRight(e.Message, "\x00")

	return err
}
