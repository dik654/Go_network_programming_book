package main

import (
	"crypto/sha512"
	"flag"
	"fmt"
	"os"
)

func init() {
	// 잘못된 플래그나 -h 또는 --help 플래그와 함께 실행될 때
	// 사용법을 나타내주는 Usage메서드 설정
	flag.Usage = func() {
		fmt.Printf("Usage: %s file...\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	// 받은 플래그 읽기
	flag.Parse()
	// 플래그를 제외한 인수들(파일명들)
	for _, file := range flag.Args() {
		// 리턴받은 sha512 체크섬과 파일명 콘솔에 뿌리기
		fmt.Printf("%s %s\n", checksum(file), file)
	}
}

func checksum(file string) string {
	// 파일을 바이너리로 읽어서
	b, err := os.ReadFile(file)
	if err != nil {
		return err.Error()
	}

	// sha512체크섬 리턴
	// 따로 콘솔(stdout)로 나가진 않음
	return fmt.Sprintf("%x", sha512.Sum512_256(b))
}
