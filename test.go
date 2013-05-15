package main

import (
	"fmt"
)

func main() {
	router := New()
	defineFib(router)
	router.Start()

	num, err := Fib(router, "", nil, 10)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Fib(10):", num)

	for key, value := range router.data {
		fmt.Println(key, "=", value)
	}
}
