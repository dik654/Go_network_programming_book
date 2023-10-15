package main

import (
	"flag"
	"log"
	"os"
	"tftp"
)

var (
	// 플래그 설정 방법 -a
	// 기본값
	// 플래그 설명
	address = flag.String("a", "127.0.0.1:69", "listen address")
	payload = flag.String("p", "payload.svg", "file to serve to clients")
)

func main() {
	// 인수로 받은 플래그 파싱
	flag.Parse()

	// ioutil이 deprecated되어서 os를 사용해야한다
	// 받은 payload주소로부터 파일을 읽어서 p변수에 저장
	p, err := os.ReadFile(*payload)
	if err != nil {
		log.Fatal(err)
	}

	// Server 인스턴스에 읽은 파일 저장
	s := tftp.Server{Payload: p}
	// udp 서버에 위에서 생성한 s를 가지고 연결
	log.Fatal(s.ListenAndServe(*address))
}
