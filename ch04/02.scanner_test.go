// bufio.Scanner를 사용하는 예
package ch04

import (
	"bufio"
	"net"
	"reflect"
	"testing"
)

const payload = "The bigger the interface, the weaker the abstraction."

func TestScanner(t *testing.T) {
	// 리스너 생성
	listener, err := net.Listen("tcp", "127.0.0.1:")
	// 리스너 생성 실패시 테스트 종료
	if err != nil {
		t.Fatal(err)
	}
	// 고루틴 생성
	go func() {
		// 리스너에서 연결 받을 준비완료
		conn, err := listener.Accept()
		if err != nil {
			// 실패시 에러 기록
			t.Error(err)
			// 후 고루틴 종료
			return
		}

		// 모든 테스트 완료 후 리스너 연결 닫기
		defer conn.Close()

		// payload 텍스트 클라이언트에 쓰기
		_, err = conn.Write([]byte(payload))
		// 쓰기 실패시 에러 기록
		if err != nil {
			t.Error(err)
		}
	}()

	// 리스너에 연결 시도
	conn, err := net.Dial("tcp", listener.Addr().String())
	// 연결 실패 시 테스트 종료
	if err != nil {
		t.Fatal(err)
	}
	// 연결 후 리스너 연결 닫기
	defer conn.Close()

	// 스캐너 선언
	scanner := bufio.NewScanner(conn)
	// 공백,구분자를 만날 때마다 데이터 분할
	scanner.Split(bufio.ScanWords)

	var words []string

	// 스캔 시작
	for scanner.Scan() {
		// scanner.Text()는 읽어들인 데이터 bytes를 문자열로 반환
		// 그리고 이를 words 배열에 추가
		words = append(words, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		t.Error(err)
	}

	// payload가 공백없이 아래와 같이 단어 배열로 들어오길 expected
	expected := []string{"The", "bigger", "the", "interface,", "the", "weaker", "the", "abstraction."}

	// expected가 실제로 들어오는지 테스트
	if !reflect.DeepEqual(words, expected) {
		// 들어오지 않는다면 테스트 종료
		t.Fatal("inaccurate scanned word list")
	}
	t.Logf("Scanned words: %#v", words)
}
