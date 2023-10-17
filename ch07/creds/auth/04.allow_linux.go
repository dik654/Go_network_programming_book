//go:build linux || darwin
// +build linux darwin

package auth

import (
	"log"
	"net"
	"os/user"

	// window 환경에서 사용불가
	"golang.org/x/sys/unix"
)

// *net.UnixConn로 유닉스 인증 정보를 알 수 있다
func Allowed(conn *net.UnixConn, groups map[string]struct{}) bool {
	if conn == nil || groups == nil || len(groups) == 0 {
		return false
	}

	// *os.File 타입 파일 디스크립터 얻기
	file, _ := conn.File()
	defer func() { _ = file.Close() }()

	var (
		err error
		// 사용자 인증 정보를 담고 있는 구조체
		// UID
		// GID
		// PID
		ucred *unix.Ucred
	)

	for {
		// file.Fd()는 파일 디스크립터
		// unix.SOL_SOCKET는 소켓 옵션 레벨
		// unix.SO_PEERCRED는 소켓 옵션 이름(연결된 peer의 UID, GID, PID 가져오기)
		ucred, err = unix.GetsockoptUcred(int(file.Fd()), unix.SOL_SOCKET, unix.SO_PEERCRED)
		// 시스템 호출이 인터럽트되었을 경우
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			log.Println(err)
			return false
		}

		break
	}

	// 사용자 정보를 조회
	// 유저명
	// UID
	// GID
	// name
	// 홈 디렉터리 정보
	u, err := user.LookupId(string(ucred.Uid))
	if err != nil {
		log.Println(err)
		return false
	}

	gids, err := u.GroupIds()
	if err != nil {
		log.Println(err)
		return false
	}
	//
	for _, gid := range gids {
		// 내가 가입한 그룹들 중에, 나에게 연결하려는 피어가 속한 그룹이 있는지 확인
		if _, ok := groups[gid]; ok {
			// 있다면 OK
			return true
		}
	}

	return false
}
