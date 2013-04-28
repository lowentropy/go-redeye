package main

import (
	"fmt"
)

func main() {
	router := New()
	defineFib(router)
	router.Start()

	value, err := router.Get("Fib", "10", "seed", "")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	num, ok := value.(int)
	if !ok {
		fmt.Println("Wrong format!")
		return
	}

	fmt.Println("Fib(10):", num, "\n")

	for key, value := range router.data {
		fmt.Println(key, "=", value)
	}
}
