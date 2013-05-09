package main

func Fib(router *Router, tgtPrefix string, tgtArgs interface{}, n int) (int, error) {
  args := [1]interface{}{n}
  value, err := router.Get("Fib", args, tgtPrefix, tgtArgs)
  return value.(int), err
}

func defineFib(__router *Router) {
  __router.Define("Fib", func(__args interface{}) (interface{}, error) {
    __args_ary, _ := __args.([1]interface{})
    n, _ := __args_ary[0].(int)

		if n < 2 {
			return 1, nil
		}
		f1, err := Fib(__router, "Fib", __args, n - 1)
		if err != nil {
			return nil, err
		}

		f2, err := Fib(__router, "Fib", __args, n - 2)
		if err != nil {
			return nil, err
		}

		return (f1 + f2), nil
  })
}
