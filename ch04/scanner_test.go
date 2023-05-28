package ch04

import (
	"bufio"
	"net"
	"reflect"
	"testing"
)

// 리스너가 이 페이로드를 제공함
const payload = "The bigger the interface, the weaker the abstraction."

func TestScanner(t *testing.T) {
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

		_, err = conn.Write([]byte(payload))
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// 네트워크 연결에서 데이터를 읽어 들일 스캐너 생성
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanWords)

	var words []string

	// 네트워크 연결에서 읽을 데이터가 있는 한 스캐너는 계속해서 데이터를 읽는다.
	// 함수를 호출할 떄마다 네트워크 연결로부터 구분자를 읽을 떄까지 여러 번의 Read 메서드를 호출하며,
	// 실패한 경우 에러를 반환
	// 네트워크 연결로부터 한 번 이상 데이터를 읽고 구분자를 찾아서 메세지를 반환하는 복잡성을 추상화한다.
	for scanner.Scan() { // io.EOF 혹은 네트워크 연결로부터 발생한 오류가 나올 떄까지 반복한다.
		// 네트워크 연결로부터 읽어 들ㅇ니 데이터 청크를 문자열로 반환
		words = append(words, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		t.Error(err)
	}

	expected := []string{"The", "bigger", "the", "interface,", "the", "weaker", "the", "abstraction."}

	if !reflect.DeepEqual(words, expected) {
		t.Fatal("inaccurate scanned word list")
	}
	//
	t.Logf("Scanned words: %#v", words)
}
