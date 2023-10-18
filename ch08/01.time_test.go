package main

import (
	"net/http"
	"testing"
	"time"
)

func TestHeadTime(t *testing.T) {
	// http://www.time.gov에 http 헤더만 받아오기
	resp, err := http.Head("http://www.time.gov")
	if err != nil {
		t.Fatal(err)
	}
	// http 연결이 아직 끊어지지 않았을 수 있으므로
	// 응답을 명시적으로 닫기
	_ = resp.Body.Close()

	// 초 단위까지 끊어서 현재 시간 now 변수에 저장
	now := time.Now().Round(time.Second)
	// 헤더의 Date 정보 읽어서 date 변수에 저장
	date := resp.Header.Get("Date")
	// date 변수가 비어있으면
	if date == "" {
		// 테스트 실패
		t.Fatal("no Date header received from time.gov")
	}

	// Mon, 02 Jan 2006 15:04:05 MST 같은 형식으로 변환
	dt, err := time.Parse(time.RFC1123, date)
	if err != nil {
		t.Fatal(err)
	}

	// 로깅
	t.Logf("time.gov: %s (skew %s)", dt, now.Sub(dt))
}
