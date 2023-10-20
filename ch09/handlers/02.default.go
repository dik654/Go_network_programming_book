package handlers

import (
	"io"
	"net/http"
	"text/template"
)

// template.New("hello")는 "hello"라는 템플릿 인스턴스를 생성
// Parse 메서드는 템플릿에 "Hello, {{.}}!" 문자열을 파싱하여 추가
// {{.}}에는 t.Execute(w, string(b))에서 string(b)에 해당하는 값이 들어감
var t = template.Must(template.New("hello").Parse("Hello, {{.}}!"))

func DefaultHandler() http.Handler {
	// 핸들러 함수를 리턴
	return http.HandlerFunc(
		// 핸들러 함수의 로직
		func(w http.ResponseWriter, r *http.Request) {
			// 핸들러 함수 종료시
			defer func(r io.ReadCloser) {
				// body 안의 데이터 모두 버리기
				_, _ = io.Copy(io.Discard, r)
				// body 닫기
				_ = r.Close()
			}(r.Body)

			var b []byte

			// 응답 메서드가 존재할 경우
			switch r.Method {
			// GET 메서드일 경우
			case http.MethodGet:
				// friend를 바이너리로 byte배열에 저장
				b = []byte("friend")
			// POST 메서드일 경우
			case http.MethodPost:
				var err error
				// 바디 전체를 byte배열에 저장
				b, err = io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			default:
				// 나머지일 경우 에러 내뿜고 종료
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			// http.ResponseWriter에 template과 string(byte 배열)를 합친 문자열에 써서 응답으로 받는다
			_ = t.Execute(w, string(b))
		},
	)
}
