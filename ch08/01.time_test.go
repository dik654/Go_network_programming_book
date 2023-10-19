package main

import (
	"net/http"
	"testing"
	"time"
)

func TestHeadTime(t *testing.T) {
	// http://www.time.gov에 http 헤더만 받아오기
	resp, err := http.Head("http://www.time.gov")
	// 아직 바디를 읽지 않은 상태여서 바디는 소비되지않고 남아있는 상태
	// 이 상태에서는 TCP 세션 재사용 불가
	if err != nil {
		t.Fatal(err)
	}
	// 응답을 명시적으로 닫기
	// 바디를 닫으면 읽지 않은 바이트를 자동으로 소비하여 TCP 세션 재사용 가능
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
