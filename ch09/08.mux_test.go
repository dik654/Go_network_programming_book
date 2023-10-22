// 요청을 특정 라우터로 라우팅(MUX 테스트)
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func drainAndClose(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// 인수로 받은 핸들러에게 요청보내고 응답 받기
			next.ServeHTTP(w, r)
			// 응답 바디 버리기
			_, _ = io.Copy(io.Discard, r.Body)
			// 인수로 받은 바디 닫기
			_ = r.Body.Close()
		},
	)
}

func TestSimpleMux(t *testing.T) {
	// 주소에 따라 다른 핸들러를 갖는 MUX 선언
	serveMux := http.NewServeMux()
	// "/"일 때 핸들러
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	// "/hello"일 때 핸들러
	serveMux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Hello friend.")
	})
	// "/hello/there/"일 떄 핸들러
	serveMux.HandleFunc("/hello/there/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Why, hello there.")
	})

	// MUX로 라우팅 요청 실행
	mux := drainAndClose(serveMux)

	testCases := []struct {
		path     string
		response string
		code     int
	}{
		{"http://test/", "", http.StatusNoContent},
		{"http://test/hello", "Hello friend.", http.StatusOK},
		{"http://test/hello/there/", "Why, hello there.", http.StatusOK},
		{"http://test/there",
			"<a href=\"/hello/there/\">Moved Permanetly</a>.\n\n",
			http.StatusMovedPermanently},
		{"http://test/hello/there/you", "Why, hello there.", http.StatusOK},
		{"http://test/hello/and/goodbye", "", http.StatusNoContent},
		{"http://test/hello/you", "", http.StatusNoContent},
	}

	for i, c := range testCases {
		// 테스트 케이스들의 경로들로 GET요청 생성
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		w := httptest.NewRecorder()
		// 실행
		mux.ServeHTTP(w, r)
		// 응답 저장
		resp := w.Result()

		// 실제 응답과 예상 비교
		if actual := resp.StatusCode; c.code != actual {
			t.Errorf("%d: expected code %d; actual %d", i, c.code, actual)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()

		if actual := string(b); c.response != actual {
			t.Errorf("%d: expected response %q; actual %q", i, c.response, actual)
		}
	}
}
