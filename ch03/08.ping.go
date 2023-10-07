// 일정한 간격으로 하트비트용 ping 전송하는 코드
package ch03

import (
	"context"
	"io"
	"time"
)

// 핑 간격 변수 선언
const defaultPingInterval = 30 * time.Second

func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	// 변수명 interval
	// 타입 time.Duration
	var interval time.Duration

	select {
	// 컨텍스트 취소된 경우 핑 함수 종료
	case <-ctx.Done():
		return
	// reset 타이머 채널에 남은 시간이 들어오면 interval로 전달하고 select 구문 넘어감
	case interval = <-reset:
	// 기본적으로 위 조건에 해당되지않아도 select 구문 넘어감
	default:
	}

	// 만약 interval로 들어온 값이 0보다 작거나 같다면
	if interval <= 0 {
		// interval을 30초로 변경
		interval = defaultPingInterval
	}

	// 30초짜리 타이머 시작
	timer := time.NewTimer(interval)
	defer func() {
		// 함수의 끝에 도달했는데 타이머가 아직 종료되지 않았다면
		if !timer.Stop() {
			// C채널에서 값이 들어올 때까지 (타이머 종료까지) 대기
			<-timer.C
		}
	}()

	for {
		select {
		// 컨텍스트가 취소되면 함수 종료
		case <-ctx.Done():
			return
		// reset 타이머로부터 남은 시간을 받으면 newInterval에 저장
		case newInterval := <-reset:
			// 아직 타이머가 종료되기 전에 reset 타이머 채널에서 시간을 받았다면
			if !timer.Stop() {
				// 종료될 때까지 대기하고 시간을 갱신할 준비
				<-timer.C
			}
			// 타이머 종료 이후 newInterval값이 0보다 크면
			if newInterval > 0 {
				// interval을 해당 남은시간(newInterval)으로 변경
				interval = newInterval
			}
		// reset 채널로 값이 들어오기 전에 Newtimer가 종료된 경우
		case <-timer.C:
			// 핑을 전송하고 종료
			if _, err := w.Write([]byte("ping")); err != nil {
				return
			}
		}
		// timer 시간을 전달받은 남은시간(newInterval := <-reset:)으로 변경
		// 시간 갱신
		_ = timer.Reset(interval)
	}
}
