package main

func defineFib(__router *Router) {
	__router.Define("Fib", func(__args interface{}) (interface{}, error) {
		var __zv int
		__args_ary, _ := __args.([1]interface{})
		n, _ := __args_ary[0].(int)
		return func() (int, error) {
			if n < 2 {
				return 1, nil
			}
			f1, __err := Fib(__router, "Fib", __args, n-1)
			if __err != nil {
				return __zv, __err
			}
			f2, __err := Fib(__router, "Fib", __args, n-2)
			if __err != nil {
				return __zv, __err
			}
			return f1 + f2, nil
		}()
	})
}

func Fib(router *Router, tgtPrefix string, tgtArgs interface{}, n int) (int, error) {
	args := [1]interface{}{n}
	value, err := router.Get("Fib", args, tgtPrefix, tgtArgs)
	return value.(int), err
}
