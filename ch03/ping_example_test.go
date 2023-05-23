package ch03

import (
	"context"
	"fmt"
	"io"
	"time"
)

func ExamplePinger() {
	ctx, cancel := context.WithCancel(context.Background())
	r, w := io.Pipe() // net.Conn 대신
	done := make(chan struct{})
	// 버퍼 채널 생성
	resetTimer := make(chan time.Duration, 1)
	resetTimer <- time.Second //초기 핑 간격

	go func() {
		Pinger(ctx, w, resetTimer)
		close(done)
	}()

	receivePing := func(d time.Duration, r io.Reader) {
		if d >= 0 {
			fmt.Printf("resetting timer (%s)\n", d)
			resetTimer <- d
		}

		now := time.Now()
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("received %q (%s)\n", buf[:n], time.Since(now).Round(100*time.Millisecond))
	}

	// 각각의 값을 receivePing에 전달
	for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
		fmt.Printf("Run %d:\n", i+1)
		receivePing(time.Duration(v)*time.Millisecond, r)
	}

	cancel()
	<-done // 콘텍스트가 취소된 이후 pinger가 종료되었는지 확인
}
