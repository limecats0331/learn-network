package ch03

import (
	"net"
	"syscall"
	"testing"
	"time"
)

// net.DialTimeout 함수가 net.Dialer 인터페이스에 대한 제어권을 제공하지 않기 때문에 테스트에서 흉내낼 수 없다.
// 따라서 동일한 인터페이스를 갖는 별도의 구현제 사용
func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{
		// net.Dialer의 Control 함수를 오버라이딩한다.
		// DNS의 타임아웃 에러를 흉내낸다.
		Control: func(_, addr string, _ syscall.RawConn) error {
			return &net.DNSError{
				Err:         "connection timed out",
				Name:        addr,
				Server:      "127.0.0.1",
				IsTimeout:   true,
				IsTemporary: true,
			}
		},
		Timeout: timeout,
	}
	return d.Dial(network, address)
}

func TestDialTimeout(t *testing.T) {
	//연결이 5초안에 성공하지 못한 경우 연결 시도는 timeout된다.
	c, err := DialTimeout("tcp", "10.0.0.1:http", 5*time.Second)
	if err == nil {
		c.Close()
		t.Fatalf("connection did not time out")
	}
	nErr, ok := err.(net.Error)
	if !ok {
		t.Fatal(err)
	}
	if !nErr.Timeout() {
		t.Fatal("error is not a timeout")
	}
}
