package ch04

import (
	"io"
	"net"
	"sync"
	"testing"
)

func proxy(from io.Reader, to io.Writer) error {
	fromWriter, fromIsWriter := from.(io.Writer)
	toReader, toIsReader := to.(io.Reader)

	if toIsReader && fromIsWriter {
		// 필요한 인터페이스는 모두 구현했다.
		// from과 to에 대응하는 프록시 생성
		go func() { _, _ = io.Copy(fromWriter, toReader) }()
	}

	_, err := io.Copy(to, from)

	return err
}

func TestProxy(t *testing.T) {
	var wg sync.WaitGroup

	// 서버는 "ping" 메세지를 대기하고 "pong" 메세지로 응답한다.
	// 그 외의 메세지는 동일하게 클라이언트로 에코잉된다.
	server, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()

				for {
					buf := make([]byte, 1024)
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}

						return
					}

					switch msg := string(buf[:n]); msg {
					case "ping":
						_, err = c.Write([]byte("pong"))
					default:
						_, err = c.Write(buf[:n])
					}

					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}

						return
					}
				}
			}(conn)
		}
	}()

	// proxyServer는 메세지를 클라이언트 연결로부터 destinationServer로 프록시한다.
	// destinationServer에서 온 응답 메세지는 역으로 클라이언트에게 프록시 된다.
	proxyServer, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// 클라이언트와의 연결이 수립되면
			conn, err := proxyServer.Accept()
			if err != nil {
				return
			}

			go func(from net.Conn) {
				defer from.Close()

				// 목적지 서버와의 연결을 수립하고
				to, err := net.Dial("tcp", server.Addr().String())
				if err != nil {
					t.Error(err)
					return
				}
				defer to.Close()

				//메세지를 프록싱한다.
				err = proxy(from, to)
				if err != nil && err != io.EOF {
					t.Error(err)
				}
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", proxyServer.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	msgs := []struct{ Message, Replay string }{
		{"ping", "pong"},
		{"pong", "pong"},
		{"echo", "echo"},
		{"ping", "pong"},
	}

	for i, m := range msgs {
		_, err = conn.Write([]byte(m.Message))
		if err != nil {
			t.Fatal(err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		if actual := string(buf[:n]); actual != m.Replay {
			t.Errorf("%d: expected reply: %q; actual: %q", i, m.Replay, actual)
		}
	}

	_ = conn.Close()
	_ = proxyServer.Close()
	_ = server.Close()
	wg.Wait()
}
