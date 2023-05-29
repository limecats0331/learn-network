package ch04

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	//메세지 타입을 나타내는 상수
	BinaryType uint8 = iota + 1
	StringType

	//보안상의 이유로 최대 페이로드 크기는 반드시 지정해야 한다.
	MaxPayloadSize uint32 = 10 << 20 // 10MB
)

var ErrMaxPayloadSzie = errors.New("maximum payload size exceeded")

type Payload interface {
	fmt.Stringer
	io.ReaderFrom //reader부터 읽는다.
	io.WriterTo   // writer에 쓴다.
	Bytes() []byte
}

// Binary 타입은 바이트 슬라이스
type Binary []byte

// 그래서 Bytes는 자기 자신을 반환
func (m Binary) Bytes() []byte {
	return m
}

// String은 string으로 캐스팅
func (m Binary) String() string {
	return string(m)
}

// io.Writer 인터페이스를 매개변수로 받아서 writer에 쓰인 바이트 수와 에러 인터페이스를 반환
func (m Binary) WriteTo(w io.Writer) (int64, error) {
	//1바이트 먼저 쓴다.
	err := binary.Write(w, binary.BigEndian, BinaryType) //1바이트 타입
	if err != nil {
		return 0, err
	}

	var n int64 = 1

	//4바이트를 쓴다.
	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4바이트 크기
	if err != nil {
		return n, err
	}
	n += 4

	//페이로드를 쓴다.
	o, err := w.Write(m) //페이로드

	return n + int64(o), err
}

func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ) //1바이트 타입
	if err != nil {
		return 0, err
	}

	var n int64 = 1
	if typ != BinaryType {
		return n, errors.New("invalid Binary")
	}

	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) //4바이트
	if err != nil {
		return n, err
	}

	n += 4
	// 페이로드 최대 사이즈는 4바이트 정수의 최대값으로 약 4GB인데
	// 악의적인 의도를 가진 사용자가 서비스 거부 공격을 시도하면 RAM을 모두 소비하기 쉽다.
	// 그래서 페이로드 최대 사이즈를 지정해줘야 한다.
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSzie
	}

	// 메세지만큼의 byte 슬라이스의 크기를 지정하고
	*m = make([]byte, size)
	// 나머지 메세지를 읽는다.
	o, err := r.Read(*m)

	return n + int64(o), err
}

type String string

func (m String) Bytes() []byte {
	return []byte(m)
}

func (m String) String() string {
	return string(m)
}

func (m String) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, StringType) // 1바이트 타입
	if err != nil {
		return 0, err
	}

	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4바이트
	if err != nil {
		return n, err
	}
	n += 4

	//writer로 쓰기전에 []byte로 형변환 해야한다.
	o, err := w.Write([]byte(m)) //페이로드

	return n + int64(o), err
}

func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ) //1
	if err != nil {
		return 0, err
	}

	var n int64 = 1
	if typ != StringType {
		return n, errors.New("invalid String")
	}

	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) // 4바이트
	if err != nil {
		return n, err
	}
	n += 4

	buf := make([]byte, size)
	o, err := r.Read(buf) // 페이로드
	if err != nil {
		return n, err
	}
	// reader로 부터 읽은 값을 String으로 형변환 한다.
	*m = String(buf)

	return n + int64(o), nil
}

func decode(r io.Reader) (Payload, error) {
	var typ uint8
	// 타입 추론을 위한 타입만큼 읽음
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}

	var payload Payload

	// 읽어 들인 타입에 따라서 상수 지정
	switch typ {
	case BinaryType:
		payload = new(Binary)
	case StringType:
		payload = new(String)
	default:
		return nil, errors.New("unknown type")
	}

	// 그냥 Raeder를 사용하면 이미 읽은 1 바이트에 대한 것을 읽을 수 없기 때문에
	// MultiReader 함수를 사용한다.
	// 지금은 읽었던 바이트를 다시 reader로 주입하는 방법이지만 첫 1바이트를 읽는 부분을 제거하는 것이 옳은 리펙터링이다.
	// TODO : MultiReader를 사용하지 않도록 리펙터링 해보기
	_, err = payload.ReadFrom(io.MultiReader(bytes.NewReader([]byte{typ}), r))
	if err != nil {
		return nil, err
	}
	return payload, nil
}
