package main

import (
	"fmt"
)

func main() {
	router := New()
	defineFib(router)
	router.Start()

	value, _ := router.Get("fib", "10", "seed", "")
	fmt.Println("Fib(10):", value.(int), "\n")

	for key, value := range router.data {
		fmt.Println(key, "=", value)
	}
}
