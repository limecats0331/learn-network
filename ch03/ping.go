package ch03

import (
	"context"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration
	select {
	case <-ctx.Done():
		return
	// 1
	case interval = <-reset: //reset 채널에서 초기 간격을 받아옴
	default:
	}

	if interval <= 0 {
		interval = defaultPingInterval
	}

	// 타이버를 interval로 초기화
	timer := time.NewTimer(interval)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	for {
		// 콘텍스트가 취소 됬거나, 타이머 리셋 시그널을 받았거나, 타이머가 만료됬거나 할때까지 블로킹된다.
		select {
		// 콘텍스트가 취소 된 경우
		case <-ctx.Done():
			return
		// 리셋을 위한 시그널을 받은 경우
		case newInterval := <-reset:
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		// 타이머가 만료되면
		case <-timer.C:
			// 핑 메세지를 쓴다.
			if _, err := w.Write([]byte("ping")); err != nil {
				return
			}
		}

		// 리셋 시그널을 받아서 다음 select 구문 실행 전에 리셋된다.
		_ = timer.Reset(interval)
	}
}
