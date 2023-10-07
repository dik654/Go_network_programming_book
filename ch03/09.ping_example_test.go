// 핑 타이머 간격 테스트
package ch03

import (
	"context"
	"fmt"
	"io"
	"time"
)

func ExamplePinger() {
	// 취소 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())
	// read, write 객체 생성
	r, w := io.Pipe()
	// done채널 생성
	done := make(chan struct{})
	// time Duration 타입을 받는 1칸짜리 resetTimer 채널 생성
	resetTimer := make(chan time.Duration, 1)
	// resetTimer 채널에 1초 값 보내기
	resetTimer <- time.Second

	// 고루틴 실행
	go func() {
		// 08에서 만든 핑 함수 실행
		Pinger(ctx, w, resetTimer)
		// 핑 함수가 끝나면 done 채널 닫기
		close(done)
	}()

	receivePing := func(d time.Duration, r io.Reader) {
		// 채널 d에 들어온 시간이 0보다 크거나 같다면
		if d >= 0 {
			fmt.Printf("resetting timer (%s)\n", d)
			// resetTimer 시간을 d로 변경
			resetTimer <- d
		}
		// 현재시간 변수에 저장
		now := time.Now()
		// 1024 버퍼 생성
		buf := make([]byte, 1024)
		// r을 읽어 버퍼에 저장하기
		// n은 저장한 바이트 수
		n, err := r.Read(buf)
		// 버퍼 읽는 중에 에러가 났다면
		if err != nil {
			// 에러 사항 콘솔에 프린트
			fmt.Println(err)
		}

		fmt.Printf("received %q (%s)\n", buf[:n], time.Since(now).Round(100*time.Millisecond))
	}

	// 실제 테스트
	for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
		fmt.Printf("Run %d:\n", i+1)
		receivePing(time.Duration(v)*time.Millisecond, r)
	}

	// 컨텍스트 취소
	cancel()
	// 컨텍스트 취소 후 pinger 함수가 종료되었는지 체크
	<-done
}
