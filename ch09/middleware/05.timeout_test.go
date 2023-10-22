// 느린 클라이언트 타임아웃 시키는 코드
package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutMiddleware(t *testing.T) {
	// 타임아웃 핸들러 선언
	handler := http.TimeoutHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 컨텐츠 없음 헤더 쓰기
			w.WriteHeader(http.StatusNoContent)
			// 1분 대기
			time.Sleep(time.Minute)
		}),
		// 핸들러가 1초안에 응답하지 못하면 타임아웃
		time.Second,
		// 타임 아웃시 바디
		"Timed out while reading response",
	)

	// GET 요청 생성
	r := httptest.NewRequest(http.MethodGet, "http://test/", nil)
	// 테스트용 응답 기록기
	w := httptest.NewRecorder()

	// 요청을 보내며 서버 실행
	handler.ServeHTTP(w, r)

	// 기록해뒀던 응답 받아서 resp 변수에 저장
	resp := w.Result()
	// 서비스 내려감 상태가 아니라면
	if resp.StatusCode != http.StatusServiceUnavailable {
		// 테스트 실패
		t.Fatalf("unexpected status code: %q", resp.Status)
	}

	// 바디 가져오기
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	// 바디 내용이 타임아웃 관련 내용이 아니라면
	if actual := string(b); actual != "Timed out while reading response" {
		// 로그 콘솔 뿌리기
		t.Logf("unexpected body: %q", actual)
	}
}
