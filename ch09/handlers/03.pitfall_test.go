package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerWriteHeader(t *testing.T) {
	// 핸들러 함수 생성
	handler := func(w http.ResponseWriter, r *http.Request) {
		// ResponseWriter에 write메서드를 사용하여 쓰면
		// 암묵적으로 상태코드는 200이 됨
		_, _ = w.Write([]byte("Bad request"))
		// 즉 이 상태코드를 400으로 쓰는 코드는 무시
		w.WriteHeader(http.StatusBadRequest)
	}

	// "http:test"로 GET 요청을 보내는 요청 생성
	r := httptest.NewRequest(http.MethodGet, "http://test", nil)
	//  테스트를 위해 실제로 응답을 보내는게 아니라 HTTP 응답을 기록
	w := httptest.NewRecorder()
	// 핸들러 함수 실행
	handler(w, r)
	// -v로 테스트 시 콘솔에 뿌리기
	t.Logf("response status: %q", w.Result().Status)

	// 핸들러 함수 선언
	handler = func(w http.ResponseWriter, r *http.Request) {
		// 이번엔 400 상태코드를 먼저 쓰고
		w.WriteHeader(http.StatusBadRequest)
		// 바디에 Bad request 쓰기
		_, _ = w.Write([]byte("Bad request"))
	}

	// GET 메서드 요청 생성
	r = httptest.NewRequest(http.MethodGet, "http://test", nil)
	//  테스트를 위해 실제로 응답을 보내는게 아니라 HTTP 응답을 기록
	w = httptest.NewRecorder()
	// 핸들러 함수 실행
	handler(w, r)
	// 테스트 콘솔에 상태코드 뿌리기
	t.Logf("Response status: %q", w.Result().Status)
}
