package ch03

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestPingerAdvanceDeadline(t *testing.T) {
	done := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	begin := time.Now()
	go func() {
		defer func() { close(done) }()
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			conn.Close()
		}()

		resetTimer := make(chan time.Duration, 1)
		resetTimer <- time.Second
		go Pinger(ctx, conn, resetTimer)

		// 데드라인의 초기값을 5초로 설정
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			t.Logf("[%s] %s",
				time.Since(begin).Truncate(time.Second), buf[:n])

			// pinger를 초기화하고
			resetTimer <- 0
			// 연결의 데드라인을 늦춘다.
			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	// 소켓이 종료되는 것을 방지하기 위해 클라이언트는 서버로부터 네 개의 핑 메세지 수신 가능
	for i := 0; i < 4; i++ { // read up to four pings
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	// 퐁 메세지 수신
	_, err = conn.Write([]byte("PONG!!!")) // should reset the ping timer
	if err != nil {
		t.Fatal(err)
	}
	// 이후 클라이언트는 네 개의 핑 메세지를 더 수신하고 데드라인이 자나기를 기다린다.
	for i := 0; i < 4; i++ { // read up to four more pings
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Fatal(err)
			}
			break
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	<-done
	end := time.Since(begin).Truncate(time.Second)
	t.Logf("[%s] done", end)
	// 서버가 연결을 끝난 시점에서 9초를 기다렸다는 것을 확인할 수 있다.
	if end != 9*time.Second {
		t.Fatalf("expected EOF at 9 seconds; actual %s", end)
	}
}
