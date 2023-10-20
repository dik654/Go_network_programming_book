package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/awoodbeck/gnp/ch09/handlers"
)

func TestSimpleHTTPServer(t *testing.T) {
	// http 서버 객체 생성
	srv := &http.Server{
		// 서버 주소 및 포트
		Addr: "127.0.0.1:8081",
		// 입출력 핸들러 함수 및 대기 제한시간 2분이 지나면
		// 3번째 인수로 넣은 에러 메세지를 바디로 하여 503 에러 리턴
		Handler: http.TimeoutHandler(
			handlers.DefaultHandler(), 2*time.Minute, ""),
		// 작업이 없는 동안 대기시간 5분
		IdleTimeout: 5 * time.Minute,
		// 서버로 보내는 요청헤더를 읽는데 1분 제한
		// 서버에 느리게 데이터를 보내 서버 연결을 점유하는 slowloris 공격 방지
		ReadHeaderTimeout: time.Minute,
	}

	// tcp 서버 리스너 생성
	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}

	// 고루틴 생성
	go func() {
		// 객체와 리스너를 바탕으로 실제 http 서버 시작
		err := srv.Serve(l)
		if err != http.ErrServerClosed {
			t.Error(err)
		}
	}()

	// 테스트 해볼 케이스들
	testCases := []struct {
		method   string
		body     io.Reader
		code     int
		response string
	}{
		//
		{http.MethodGet, nil, http.StatusOK, "Hello, friend!"},
		{http.MethodPost, bytes.NewBufferString("<world>"), http.StatusOK, "Hello, &lt;world&gt;!"},
		{http.MethodHead, nil, http.StatusMethodNotAllowed, ""},
	}

	// http 클라이언트 생성
	client := new(http.Client)
	// "http://서버 아이피/"로 연결할 경로 변수 생성
	path := fmt.Sprintf("http://%s/", srv.Addr)

	for i, c := range testCases {
		// 테스트 케이스의 메서드로 서버 아이피로 바디와 함께 요청 생성
		r, err := http.NewRequest(c.method, path, c.body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			// 요청 중에 에러가 있어도 다음 케이스 실행
			continue
		}

		// 위에서 만든 요청을 가지고
		// 테스트 서버에 실제로 요청을 날린 뒤
		// 응답을 받아 resp 변수에 저장
		resp, err := client.Do(r)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		// 상태코드가 테스트 케이스의 것과 다르면
		if resp.StatusCode != c.code {
			// 에러 뿌리기
			t.Errorf("%d: unexpected status code: %q", i, resp.Status)
		}

		// 응답 바디를 모두 읽은 뒤
		// 변수 b에 저장하고
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		// 바디 닫기
		_ = resp.Body.Close()

		// 테스트 케이스의 응답과 실제 응답이 다르면
		if c.response != string(b) {
			// 에러 뿌리기
			t.Errorf("%d: expected %q; actual %q", i, c.response, b)
		}
	}

	// 테스트가 모두 끝났으니 서버 닫기
	if err := srv.Close(); err != nil {
		// 서버 닫기 실패 시 테스트 실패
		t.Fatal(err)
	}
}
