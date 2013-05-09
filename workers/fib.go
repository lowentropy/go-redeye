package main

type wat struct {
	x int
}

func Fib(n *wat) (*wat, error) {
	if n.x < 2 {
		return &wat{1}, nil
	}
	f1 := Fib(&wat{n.x - 1})
	f2 := Fib(&wat{n.x - 2})
	return &wat{f1.x + f2.x}, nil
}
