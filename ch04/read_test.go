package ch04

import (
	"io"
	"math/rand"
	"net"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
	// 클라이언트가 읽어 드릴 16MB 페이로드의 랜덤 데이터를 생성
	payload := make([]byte, 1<<24) // 16MB
	_, err := rand.Read(payload)
	if err != nil {
		t.Fatal(err)
	}

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
		defer conn.Close()

		// 연결을 수신하고 나서 서버는 네트워크 연결로 페이로드 전체를 쓴다.
		_, err = conn.Write(payload)
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	// 클라이언트 버퍼의 크기 512kb
	buf := make([]byte, 1<<19) //512kb

	// 읽는 데이터 16MB보다 버퍼가 작음으로 여려번 읽어야 한다.
	for {
		// 다음 순회전까지 연결로부터 512KB를 읽는다.
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			break
		}
		t.Logf("read %d bytes", n)
	}

	conn.Close()
}
