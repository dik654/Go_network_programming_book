package handlers

import (
	"io"
	"net/http"
	"sort"
	"strings"
)

// 조건에 따라 핸들러 변경하기 위해 매핑
type Methods map[string]http.Handler

func (h Methods) serveHTTP(w http.ResponseWriter, r *http.Request) {
	// 함수 종료 시 바디를 비우고 바디를 닫아 http를 재사용할 수 있게함
	defer func(r io.ReadCloser) {
		_, _ = io.Copy(io.Discard, r)
		_ = r.Close()
	}(r.Body)

	// GET, POST등 요청에 맞춰서 handler변수에 핸들러 함수 저장
	if handler, ok := h[r.Method]; ok {
		// 해당 핸들러 함수가 없는 경우
		if handler == nil {
			// 에러 뿌리기
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		} else {
			// 있다면 요청 처리
			handler.ServeHTTP(w, r)
		}
		return
	}

	//
	w.Header().Add("Allow", h.allowedMethods())
	if r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h Methods) allowedMethods() string {
	// 현재 비어있고 최대 공간이 h인 string 슬라이스 생성
	a := make([]string, 0, len(h))

	// 메서드들을 돌면서
	for k := range h {
		// 슬라이스에 모두 추가
		a = append(a, k)
	}
	// a, b, c.. 순으로 재배열
	// a = []string{"DELETE", "GET", "POST"}
	sort.Strings(a)

	// 슬라이스의 각각의 원소들을 , 을 구분자로 하나의 문자열로 합치기
	// a = "DELETE, GET, POST"
	return strings.Join(a, ", ")
}
