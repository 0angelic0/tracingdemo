package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		panic("ERROR: Expecting one argument")
	}
	helloTo := os.Args[1]
	helloStr := fmt.Sprintf("Hello, %s!", helloTo)
	println(helloStr)
}
