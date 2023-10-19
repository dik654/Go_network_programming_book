package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type User struct {
	First string
	Last  string
}

func handlePostUser(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// 테스트 종료 시 io.Reader와 io.Closer가 합쳐져있는 io.ReadCloser 사용
		defer func(r io.ReadCloser) {
			// ioutil.Discard는 Deprecated 되었음
			// r에서 읽은 데이터를 io.Discard에 복사
			// 즉, 데이터를 버림
			_, _ = io.Copy(io.Discard, r)
			//
			_ = r.Close()
		}(r.Body)

		// Request가 POST 요청이 아니라면
		if r.Method != http.MethodPost {
			// 메서드 허가X 상태 에러를 나타내며
			http.Error(w, "", http.StatusMethodNotAllowed)
			// 함수 종료
			return
		}

		var u User
		// 바디를 받아서 JSON 디코더 생성
		// 디코더는 JSON으로 디코딩하여 u주소에 저장
		err := json.NewDecoder(r.Body).Decode(&u)
		// 디코딩 실패 시
		if err != nil {
			// 에러 이유
			t.Error(err)
			// 잘못된 요청 에러
			http.Error(w, "Decode Failed", http.StatusBadRequest)
			// 함수 종료
			return
		}

		// 헤더에 상태 허가 쓰기
		w.WriteHeader(http.StatusAccepted)
	}
}

func TestPostUser(t *testing.T) {
	// 테스트 서버에서 위의 만든 함수를 통해 요청을 처리
	ts := httptest.NewServer(http.HandlerFunc(handlePostUser(t)))
	// 테스트 종료시 서버 닫기
	defer ts.Close()

	// 테스트 서버로 GET요청
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	// 상태코드가 허가되지않은 메서드가 아니라면
	if resp.StatusCode != http.StatusMethodNotAllowed {
		// 테스트 실패
		t.Fatalf("expected status %d; actual status %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}

	// 버퍼 생성
	buf := new(bytes.Buffer)
	u := User{First: "Adam", Last: "Woodbeck"}
	// User 객체를 인코딩하여 버퍼에 쓰기
	err = json.NewEncoder(buf).Encode(&u)
	if err != nil {
		t.Fatal(err)
	}

	// 테스트 서버로 application/json 형식으로 버퍼에 있는 데이터를 POST하기를 요청
	resp, err = http.Post(ts.URL, "application/json", buf)
	if err != nil {
		t.Fatal(err)
	}

	// 상태코드가 "허가"가 아니라면
	if resp.StatusCode != http.StatusAccepted {
		// 테스트 실패
		t.Fatalf("expected status %d; actual status %d", http.StatusAccepted, resp.StatusCode)
	}
	// 바디 닫기
	_ = resp.Body.Close()
}

func TestMultipartPost(t *testing.T) {
	// 버퍼 생성
	reqBody := new(bytes.Buffer)
	// multipart는 다른 종류의 데이터를 하나의 요청에 담을 수 있도록 해줌
	w := multipart.NewWriter(reqBody)

	// map에 시간 데이터와 텍스트 데이터 저장
	for k, v := range map[string]string{
		"date":        time.Now().Format(time.RFC3339),
		"description": "Form values with attached files",
	} {
		// multipart writer로 간 데이터와 텍스트 데이터 쓰기
		err := w.WriteField(k, v)
		if err != nil {
			t.Fatal(err)
		}
	}

	// 60초 타임아웃 컨텍스트 생성
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 컨텍스트를 포함한 POST 요청 생성
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://httpbin.org/post", reqBody)
	if err != nil {
		t.Fatal(err)
	}
	// 요청 헤더에 Content-Type 설정
	req.Header.Set("Content-Type", w.FormDataContentType())

	// 요청 보내기
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	// 테스트 종료 시 바디 닫기
	defer func() { _ = resp.Body.Close() }()

	// ioutil.ReadAll은 deprecated
	// 바디 읽어서 변수 b에 저장
	b, err := io.ReadAll(resp.Body)

	if err != nil {
		t.Fatal(err)
	}

	// 상태코드가 OK가 아니라면
	if resp.StatusCode != http.StatusOK {
		// 테스트 실패
		t.Fatalf("expected status %d; actual %d", http.StatusOK, resp.StatusCode)
	}

	t.Logf("\n%s", b)
}
