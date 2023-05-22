package ch03

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestDeadline(t *testing.T) {
	sync := make(chan struct{})

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer func() {
			conn.Close()
			close(sync)
		}()

		// TCP 연결을 수락하고 나서 연결살의 읽기에 대한
		// 데스트상에서 클라이언트가 데이터를 전송하지 않기 때문에
		// Read 메서드는 데드라인이 지날 때까지 블로킹 될 것이다.
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		nErr, ok := err.(net.Error)
		// Read 메섣드는 타임아웃으로 설정한 5초 뒤에 에러를 반환한다.
		if !ok || !nErr.Timeout() {
			t.Errorf("expected timeout error; actual: %v", err)
		}

		sync <- struct{}{}

		// 데드라인을 좀 더 뒤로 설정
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	<-sync
	_, err = conn.Write([]byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	// Read에서 블록킹된 메서드는 EOF를 반환받는다.
	if err != io.EOF {
		t.Errorf("expected server termination; actual: %v", err)
	}
}
