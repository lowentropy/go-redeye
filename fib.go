package main

type wat struct {
	x int
}

func defineFib(__router *Router) {;__router.Define("Fib", func(__args interface{}) (interface{}, error) {;__args_ary, _ := __args.([1]interface{});n, _ := __args_ary[0].(*wat);return func() (*wat, error) {
	if n.x < 2 {
		return &wat{1}, nil
	}
	f1, err := Fib(__router, "Fib", __args, &wat{n.x - 1}); if err != nil { return nil, err }
	f2, err := Fib(__router, "Fib", __args, &wat{n.x - 2}); if err != nil { return nil, err }
	return &wat{f1.x + f2.x}, nil
}()})}

            func Fib(router *Router, tgtPrefix string, tgtArgs interface{}, n *wat) (*wat, error) {
              args := [1]interface{}{n}
              value, err := router.Get("Fib", args, tgtPrefix, tgtArgs)
              return value.(*wat), err
            }
