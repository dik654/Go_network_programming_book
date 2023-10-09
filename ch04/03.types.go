// 가변 크기 데이터 처리를 위해 TLV 인코딩 데이터 유형 설정
package ch04

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	BinaryType uint8 = iota + 1
	StringType

	MaxPayloadSize uint32 = 10 << 20
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

// Binary 타입 생성
type Binary []byte

func (m Binary) Bytes() []byte  { return m }
func (m Binary) String() string { return string(m) }

// Writer에 쓰기 메서드
func (m Binary) WriteTo(w io.Writer) (int64, error) {
	// BinaryType uint8 = iota + 1을 writer에 쓰기
	err := binary.Write(w, binary.BigEndian, BinaryType)
	if err != nil {
		return 0, err
	}
	// 1byte를 썼음을 표기
	var n int64 = 1
	// 32bit = 4bytes
	// Binary 인스턴스의 길이 4bytes를 writer에 쓰기
	err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}
	// 추가로 4bytes를 썼음을 표기
	n += 4
	// Binary 인스턴스자체를 writer에 쓰기
	o, err := w.Write(m)
	// 5bytes + 인스턴스의 bytes크기, 에러상태 리턴
	return n + int64(o), err
}

// Reader로 읽어오기 메서드
func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	// writer로부터 타입 1byte 읽어오기
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return 0, err
	}
	// 1byte 읽었다는 표시
	var n int64 = 1
	// 타입이 BinaryType이 아니라면
	if typ != BinaryType {
		// 1, 유효하지않은 Binary 에러 리턴
		return n, errors.New("Invalid Binary")
	}
	// 4bytes Binary 인스턴스 길이
	var size uint32
	// 이미 읽은 1byte 이후 4bytes 가져오기
	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return n, err
	}
	// 읽은 기록 갱신
	n += 4
	// 만약 읽은 인스턴스가 길이가 제한길이를 넘어서면 에러 리턴하고 메서드 종료
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}

	// 인스턴스 길이에 맞춰 버퍼 생성
	*m = make([]byte, size)
	// 인스턴스 값 가져오기
	o, err := r.Read(*m)
	// 읽은 전체 길이 리턴
	return n + int64(o), err
}

// String 타입 생성
type String string

func (m String) Bytes() []byte  { return []byte(m) }
func (m String) String() string { return string(m) }

// 쓰기
func (m String) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, StringType)
	if err != nil {
		return 0, err
	}

	var n int64 = 1
	err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}
	n += 4

	o, err := w.Write([]byte(m))

	return n + int64(o), err
}

// 읽기
func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	if typ != StringType {
		return n, errors.New("invalid String")
	}

	var size uint32
	err = binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return n, err
	}
	n += 4

	buf := make([]byte, size)
	o, err := r.Read(buf)
	if err != nil {
		return n, err
	}
	*m = String(buf)

	return n + int64(o), nil
}

// 읽은 데이터를 타입에 맞게 디코딩
func decode(r io.Reader) (Payload, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}

	var payload Payload

	switch typ {
	case BinaryType:
		payload = new(Binary)
	case StringType:
		payload = new(String)
	default:
		return nil, errors.New("unknown type")
	}

	_, err = payload.ReadFrom(
		io.MultiReader(bytes.NewReader([]byte{typ}), r))
	if err != nil {
		return nil, err
	}

	return payload, nil
}
