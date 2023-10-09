// 두 노드간 데이터를 전송하는 proxy 생성
package ch04

import (
	"io"
	"net"
)

func proxyConn(source, destination string) error {
	// 보내는 쪽 연결
	connSource, err := net.Dial("tcp", source)
	if err != nil {
		return err
	}
	// proxy함수가 종료되면 보내는 쪽 연결 해제
	defer connSource.Close()

	// 받는 쪽 연결
	connDestination, err := net.Dial("tcp", destination)
	if err != nil {
		return err
	}
	// proxy함수가 종료되면 받는 쪽 연결 해제
	defer connDestination.Close()

	// connSource에 데이터가 쓰일때마다 connDestination으로 복사하기
	go func() { _, _ = io.Copy(connSource, connDestination) }()
	// connDestination에 데이터가 쓰일때마다 connSource으로 복사하기
	_, err = io.Copy(connDestination, connSource)

	return err
}
