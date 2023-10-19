package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func blockIndefinitely(w http.ResponseWriter, r *http.Request) {
	// case가 없어 무한정 대기
	select {}
}

// 무한정 대기하는지 테스트
func TestBlockIndefinitely(t *testing.T) {
	// 테스트 http 서버를 생성하여
	// blockIndefinitely로 요청을 넘김
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
	// 테스트 서버에 GET요청
	_, _ = http.Get(ts.URL)
	// 이 라인에 도달하면 테스트 실패
	t.Fatal("client did not indefinitely block")
}

func TestBlockIndefinitelyWithTimeout(t *testing.T) {
	// 테스트 http 서버를 생성하여
	// blockIndefinitely로 요청을 넘김
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
	// 타임아웃 컨텍스트 생성
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 테스트 종료 시 컨텍스트 취소
	defer cancel()

	// 컨텍스트와 함께 GET 요청을 하는 req 생성
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// 테스트 서버에 요청 보내기
	resp, err := http.DefaultClient.Do(req)
	// 요청 보내는 도중 에러가 생겼다면
	if err != nil {
		// 그리고 에러가 타임아웃 에러가 아니라면
		if !errors.Is(err, context.DeadlineExceeded) {
			// 테스트 실패
			t.Fatal(err)
		}
		// 타임아웃이면 테스트 종료
		return
	}

	// 응답 바디 닫기
	_ = resp.Body.Close()
}
