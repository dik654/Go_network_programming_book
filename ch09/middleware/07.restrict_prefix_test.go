package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRestrictPrefix(t *testing.T) {
	// StripPrefix는 주소에서 /static/을 제외한 나머지만 리턴
	// RestrictPrefix에서 주소에 .가 있는지 확인
	// FileServer는 GET요청을 받으면 ../files/내의 파일을 리턴
	handler := http.StripPrefix(
		"/static/",
		RestrictPrefix(".", http.FileServer(http.Dir("../files/"))),
	)

	testCases := []struct {
		path string
		code int
	}{
		{"http://test/static/sage.svg", http.StatusOK},
		{"http://test/static/.secret", http.StatusNotFound},
		{"http://test/static/.dir/secret", http.StatusNotFound},
	}

	for i, c := range testCases {
		// GET 요청 생성
		r := httptest.NewRequest(http.MethodGet, c.path, nil)
		// 응답 저장기 선언
		w := httptest.NewRecorder()
		// 핸들러로 GET요청 보내기
		handler.ServeHTTP(w, r)

		// 실제 결과 상태코드가 예상과 다르면
		actual := w.Result().StatusCode
		if c.code != actual {
			// 에러 콘솔에 뿌리기
			t.Errorf("%d: expected %d; actual %d", i, c.code, actual)
		}
	}
}
