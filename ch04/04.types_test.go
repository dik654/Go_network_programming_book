// 03 types 테스트하기
package ch04

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

func TestPayloads(t *testing.T) {
	b1 := Binary("Clear is better than clever.")
	b2 := Binary("Don't panic.")
	s1 := String("Errors are values.")
	payloads := []Payload{&b1, &s1, &b2}
	// 리스너 생성
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// 고루틴 실행
	go func() {
		// 연결 받을 준비
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		// 고루틴 종료시 연결 끊기
		defer conn.Close()
		// payload 순서대로 리스너에 쓰기
		for _, p := range payloads {
			_, err = p.WriteTo(conn)
			if err != nil {
				t.Error(err)
				break
			}
		}
	}()
	// 리스너 연결 시도
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for i := 0; i < len(payloads); i++ {
		// 쓴 데이터 차례대로 디코딩
		actual, err := decode(conn)
		if err != nil {
			t.Fatal(err)
		}

		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			t.Errorf("value mismatch: %v != %v", expected, actual)
			continue
		}

		t.Logf("[%T] %[1]q", actual)
	}
}

func TestMaxPayloadSize(t *testing.T) {
	// io.Reader, io.Writer, ReaderFrom, WriterTo 인테퍼이스가 구현된 버퍼 생성
	buf := new(bytes.Buffer)
	// buf에 타입 정보 쓰기
	err := buf.WriteByte(BinaryType)
	if err != nil {
		t.Fatal(err)
	}
	// buf에 1GB 넣기
	err = binary.Write(buf, binary.BigEndian, uint32(1<<30))
	if err != nil {
		t.Fatal(err)
	}

	var b Binary
	// buf를 읽어서
	_, err = b.ReadFrom(buf)
	// 최대 크기 에러가 났는지 확인
	if err != ErrMaxPayloadSize {
		// 안났다면 테스트 에러 종료
		t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
	}
}
