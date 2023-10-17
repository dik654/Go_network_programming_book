package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"

	"github.com/awoodbeck/gnp/ch07/creds/auth"
)

// 초기화 함수
func init() {
	// 잘못된 플래그나 help시 보여주는 내용
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n\t%s <group names>\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func parseGroupNames(args []string) map[string]struct{} {
	// 그룹 맵 생성
	groups := make(map[string]struct{})

	for _, arg := range args {
		// 내 그룹명 가져와서
		grp, err := user.LookupGroup(arg)
		if err != nil {
			log.Println(err)
			continue
		}
		// 매핑
		groups[grp.Gid] = struct{}{}
	}

	return groups
}

func main() {
	// 플래그 파싱
	flag.Parse()

	// 받은 플래그에서 그룹명 매핑
	groups := parseGroupNames(flag.Args())
	socket := filepath.Join(os.TempDir(), "creds.sock")
	// net.UnixAddr 구조체의 포인터를 반환
	// Unix 도메인 소켓을 생성하거나 연결할 때 사용
	addr, err := net.ResolveUnixAddr("unix", socket)
	if err != nil {
		log.Fatal(err)
	}

	// 리스너 생성
	s, err := net.ListenUnix("unix", addr)
	if err != nil {
		log.Fatal(err)
	}

	// os.Signal 타입의 채널을 생성
	c := make(chan os.Signal, 1)
	// ctrl + c를 눌렀을 때 인터럽트를 c채널로 전달
	signal.Notify(c, os.Interrupt)

	// 고루틴 생성
	go func() {
		// c채널로 값이 들어오면
		<-c
		// 리스너 닫기
		_ = s.Close()
	}()

	fmt.Printf("Listening on %s ...\n", socket)

	for {
		// 연결 수락 준비완료
		conn, err := s.AcceptUnix()
		if err != nil {
			break
		}
		// 연결하려는 피어와 같은 그룹이 존재하면 TRUE
		if auth.Allowed(conn, groups) {
			// 리스너에 Welcome 쓰기
			_, err = conn.Write([]byte("Welcome\n"))
			if err == nil {
				continue
			}
		} else {
			// 아니면 연결 실패 쓰기
			_, err = conn.Write([]byte("Access denied\n"))
		}
		if err != nil {
			log.Println(err)
		}
		// 연결 끊기
		_ = conn.Close()
	}
}
