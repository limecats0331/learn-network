package ch03

import (
	"context"
	"log"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
	dl := time.Now().Add(5 * time.Second)                         //현재 시간으로부터 5초 뒤 시간 저장
	ctx, cancel := context.WithDeadline(context.Background(), dl) // 데드라인 설정
	defer cancel()                                                // gc되도록 cancel 함수 defer로 호출

	var d net.Dialer
	//Control을 override해서 deadline보다 조금 넘는 시간동안 지연
	d.Control = func(_, _ string, _ syscall.RawConn) error {
		time.Sleep(5*time.Second + time.Millisecond)
		return nil
	}
	// DailContext에 생성한 Context를 전달
	conn, err := d.DialContext(ctx, "tcp", "naver.com:https")
	if err == nil {
		conn.Close()
		log.Printf("err=%v\n", err)
		t.Fatal("connection did not time out")
	}
	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			log.Printf("err=%v\n", err)
			t.Errorf("error is not a timeout: %v", err)
		}
	}
	//데드라인이 콘텍스트를 제대로 취소 했는지, cancel 함수 호출은 문제 없는지 확인
	//데드라인에 의해 취소되었는지?
	if ctx.Err() != context.DeadlineExceeded {
		log.Printf("ctx=%v\n", ctx.Err())
		t.Errorf("expected deadline exceeded; acutal: %v", ctx.Err())
	}
}
