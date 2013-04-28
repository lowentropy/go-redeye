package main

import "fmt"

func Fib(router *Router, tgtPrefix, tgtArgs string, n int) (int, error) {
	args := fmt.Sprintf("%v", n)
	value, err := router.Get("Fib", args, tgtPrefix, tgtArgs)
	return value.(int), err
}

func defineFib(__router *Router) {
	__router.Define("Fib", func(__args string) (interface{}, error) {
		var n int
		fmt.Sscanf(__args, "%v", &n)

		if n < 2 {
			return 1, nil
		}
		f1, _ := Fib(__router, "Fib", __args, n - 1)
		f2, _ := Fib(__router, "Fib", __args, n - 2)
		return (f1 + f2), nil
	})
}

