package main

func Fib(n int) (int, error) {
	if n < 2 {
		return 1, nil
	}
	f1 := Fib(n - 1)
	f2 := Fib(n - 2)
	return (f1 + f2), nil
}
