package ch04

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

func TestPayloads(t *testing.T) {
	// 2개의 바이너리 타입과 2개의 스트링타입을 생성
	b1 := Binary("Clear is better than clever.")
	b2 := Binary("Don't panic.")
	s1 := String("Errors are values.")
	// 이것을 Payload 인터페이스의 슬라이스에 등록
	payloads := []Payload{&b1, &b2, &s1}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		// 연결을 수립할 리스너에 각 타입을 전송
		for _, p := range payloads {
			_, err = p.WriteTo(conn)
			if err != nil {
				t.Error(err)
				break
			}
		}
	}()

	// 리스너로 연결을 수립한다.
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// 페이로드 슬라이스 크기 만큼 받음
	for i := 0; i < len(payloads); i++ {
		// 각 페이로드를 디코딩
		actual, err := decode(conn)
		if err != nil {
			t.Fatal(err)
		}

		// 디코딩된 타입을 서버가 전송한 타입과 비교한다.
		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			// 만약 변수의 타입이나 페이로드가 비교한 것과 다르면 실패한다.
			t.Errorf("value mismatch: %v != %v", expected, actual)
			continue
		}

		t.Logf("[%T] %[1]q", actual)
	}
}

func TestMaxPayloadSize(t *testing.T) {
	buf := new(bytes.Buffer)
	err := buf.WriteByte(BinaryType)
	if err != nil {
		t.Fatal(err)
	}

	// 페이로드의 최대 크기가 1GB
	err = binary.Write(buf, binary.BigEndian, uint32(1<<30))
	if err != nil {
		t.Fatal(err)
	}

	var b Binary
	_, err = b.ReadFrom(buf)
	if err != ErrMaxPayloadSzie {
		t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
	}
}
