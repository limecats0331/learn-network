package main

import "fmt"

type A interface {
	A() bool
}

type B interface {
	B() bool
}

type C struct {
	name string
}

func (c *C) A() bool {
	return true
}

func (c *C) B() bool {
	return true
}

func isInputB(b B) bool {
	return true
}

func main() {
	CImpl := new(C)

	fmt.Println(isInputB(CImpl))
}
