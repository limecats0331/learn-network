package ch04

import (
	"io"
	"net"
)

func proxyConn(source, destination string) error {
	// 출발지 노드와 연결을 생성
	connSource, err := net.Dial("tcp", source)
	if err != nil {
		return err
	}
	defer connSource.Close()

	// 목적지 노드와 연결을 생성
	connDestination, err := net.Dial("tcp", destination)
	if err != nil {
		return err
	}
	defer connDestination.Close()

	go func() {
		// connSource에 대응하는 connDestination
		// io.Writer와 io.Reader의 인터페이스를 받아서 reader로부터 읽은 모든 데이터를 writer로 전송
		// connDestination으로부터 데이터를 읽고 connSource로 데이터를 쓴다.
		_, _ = io.Copy(connSource, connDestination)
	}()

	// connDestination으로 메세지를 보내는 connSource
	_, err = io.Copy(connDestination, connSource)

	return err
}
