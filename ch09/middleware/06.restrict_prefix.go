// 특정 경로 금지용
package middleware

import (
	"net/http"
	"path"
	"strings"
)

func RestrictPrefix(prefix string, next http.Handler) http.Handler {
	return http.HandlerFunc(
		//
		func(w http.ResponseWriter, r *http.Request) {
			// "/"를 구분자로 주소의 문자열을 나눠서
			for _, p := range strings.Split(path.Clean(r.URL.Path), "/") {
				// 인수의 prefix에 해당하는 문자열이 있다면
				if strings.HasPrefix(p, prefix) {
					// 404를 뿌리고
					http.Error(w, "Not Found", http.StatusNotFound)
					// 함수 종료
					return
				}
			}
			// 아니라면 요청 실행
			next.ServeHTTP(w, r)
		},
	)
}
