package main

func Fib(n int) (int, error) {
	if n < 2 {
		return 1, nil
	}
	f1, _ := Fib(n - 1)
	f2, _ := Fib(n - 2)
	return (f1 + f2), nil
}
