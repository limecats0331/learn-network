package ch03

import (
	"io"
	"log"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// Create a listener on a random port.
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{ string })
	go func() { //리스너의 고루틴
		defer func() { done <- struct{ string }{"listener end"} }()

		for {
			//바로 블로킹이 해제되며 에러를 반환
			//무언가 실패했다는 의미가 아니므로 그냥 로깅하고 넘어가면 된다.
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			go func(c net.Conn) {
				defer func() {
					c.Close() //FIN 패킷 전송하면서 TCP 연결 마무리
					done <- struct{ string }{"connection done"}
				}()

				buf := make([]byte, 1024)
				for {
					// FIN 패킷을 받으면 Read 함수는 EOF 오류를 반환한다.
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF { //반대편 연결이 종료되었다.
							t.Error(err)
						}
						return
					}

					t.Logf("received: %q", buf[:n])
				}
			}(conn) // <- 익명함수 파라미터
		}
	}()

	//tcp 연결을 시도하는 클라이언트 부분
	//어떤 프로토콜인지, IP 주소나 호스트 이름 사용 가능
	//IPv6는 구분자로 클론을 구분자로 사용하기 때문에 대괄호로 감싸야 한다. [2001:ed27::1]:https
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	conn.Close() //우아한 종료 시작
	log.Printf("%v\n", <-done)
	listener.Close()
	log.Printf("%v\n", <-done)
	//<-done
}
