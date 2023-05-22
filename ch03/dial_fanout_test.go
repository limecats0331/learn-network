package ch03

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextCancelFanout(t *testing.T) {
	// Err()는 Cancel, DeadlineExceeded, nil 셋 중에 하나의 값을 반환
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(10*time.Second),
	)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	//리스너는 하나의 연결을 수락하고 성공적으로 핸드세이크를 마치면 연결을 종료한다.
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	// 여러 개의 다이얼러를 실행하기 때문에
	// 다이얼링을 위한 코드를 추상화하여 별도의 함수로 분리한다.
	dial := func(ctx context.Context, address string, response chan int,
		id int, wg *sync.WaitGroup) {
		defer wg.Done()

		var d net.Dialer
		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}
		c.Close()

		select {
		case <-ctx.Done():
		case response <- id:
		}
	}

	res := make(chan int)
	var wg sync.WaitGroup

	// 별도의 고루틴을 호출하여 여러 개의 다이얼러를 생성
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), res, i+1, &wg)
	}

	// 정상적으로 동작하면 한 연결 시도는 다른 연결 시도보다 먼저 성공적으로 리스너에 연결될 수 있다.
	// 연결에 성공한 다이얼러의 ID를 res 채널에서 받는다.
	response := <-res
	cancel()
	// 다이얼러들의 연결 시도 중단이 끝나고 고루틴이 중료될때까지 블로킹한다.
	wg.Wait()
	close(res)

	// 콘텍스트 취소가 코드상의 취소였음을 확인
	//cancel이 호출되었더라도 데드라인이 지나서 콘텍스트가 취소될 수도 있음
	//그러면 context.DeadlineExceeded를 반환한다.
	if ctx.Err() != context.Canceled {
		t.Errorf("expected canceled context; actual: %s", ctx.Err())
	}

	t.Logf("dialer %d retrieved the resource", response)
}
