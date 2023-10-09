// 특정 포트와 TCP 연결하는데 걸리는 시간을 체크하는 코드
package ch04

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var (
	// 플래그 이름 "c", --c로 호출
	// 기본 값 "3"
	// 뒤의 문자열은 사용법
	// 반환 값은 해당 플래그의 값을 저장할 정수의 포인터(*int)
	count = flag.Int("c", 3, "number of pings: <= means forever")
	// 플래그 이름 "i"
	// 기본 값 1초
	// Duration 값을 변수에 저장
	interval = flag.Duration("i", time.Second, "interval between pings")
	timeout  = flag.Duration("W", 5*time.Second, "timeto wait for a reply")
)

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] host:port\nOptions:\n", os.Args[0])
		// 모든 플래그와 그 사용법을 출력
		flag.PrintDefaults()
	}
}

func main() {
	// 명령행 인수를 파싱
	flag.Parse()

	// NArgs는 플래그가 아닌 인수의 개수를 반환
	// 플래그가 아닌 인수가 1개가 아니라면
	if flag.NArg() != 1 {
		fmt.Print("host:port is required\n\n")
		// 플래그 사용법 보여줌
		flag.Usage()
		os.Exit(1)
	}

	// 플래그가 아닌 첫 번째 인수(host:port)를 반환
	target := flag.Arg(0)
	fmt.Println("PING", target)

	// 인수로 들어온 count 값이 0보다 작으면
	if *count <= 0 {
		// CTRL C로 끌 수 있음을 알림
		fmt.Println("CTRL+C to stop.")
	}

	// count를 세기 위한 msg 변수 0으로 선언
	msg := 0

	// count가 0보다 작거나 같아서 무한반복이거나
	// msg 값이 count보다 작을 경우 반복
	for (*count <= 0) || (msg < *count) {
		// 횟수++
		msg++
		fmt.Print(msg, " ")

		// 시작 시간 저장
		start := time.Now()
		// 리스너 연결 요청
		c, err := net.DialTimeout("tcp", target, *timeout)
		// 연결까지 걸린 시간 저장
		dur := time.Since(start)

		// 연결 실패 시
		if err != nil {
			// 실패 사유 프린트
			fmt.Printf("fail in %s: %v\n", dur, err)
			if nErr, ok := err.(net.Error); !ok || !nErr.Temporary() {
				os.Exit(1)
			}
		} else {
			// 정상 연결에 성공한 경우 연결 해제
			_ = c.Close()
			// 연결까지 걸린시간 프린트
			fmt.Println(dur)
		}

		// 다음 반복까지 interval 시간동안 대기
		time.Sleep(*interval)
	}
}
