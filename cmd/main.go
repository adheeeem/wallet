package main

import (
	"fmt"
	"time"
)

func longOperation(i int) {
	time.Sleep(time.Second)
	fmt.Printf("%v passed", i)
}

func main() {

	for i := 0; i < 5; i++ {
		go longOperation(i)
	}
	fmt.Println("that's all")
}
