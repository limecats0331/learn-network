package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContextCancel(t *testing.T) {
	//콘텍스트와 콘텍스트를 취소할 수 있는 함수를 같이 받는다.
	ctx, cancel := context.WithCancel(context.Background())
	sync := make(chan struct{})

	//수동으로 콘텍스트를 취소하기 때문에 클로저를 만들어서 별도로 연결 시도를 처리하기 위한 고루틴 시작
	go func() {
		defer func() { sync <- struct{}{} }()

		var d net.Dialer
		d.Control = func(_, _ string, _ syscall.RawConn) error {
			time.Sleep(time.Second)
			return nil
		}

		conn, err := d.DialContext(ctx, "tcp", "127.0.0.1:80")
		if err != nil {
			t.Log(err)
			return
		}

		conn.Close()
		t.Error("connection did not time out")

	}()

	//원격 노드와 핸드쉐이크가 끝나면 콘텍스트를 취소하기 위한 cancel 실행
	cancel()
	<-sync

	//여기서 콘텍스트는 nil이 아닌 에러를 반환하면서 종료하는데
	//cancel호출로 반환된 에러는 context.Canceled여야 한다.
	if ctx.Err() != context.Canceled {
		t.Errorf("expected canceled context; actual: %q", ctx.Err())
	}
}
