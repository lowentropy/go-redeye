# go-redeye

This is a quick attempt at an all-Go version of [redeye](https://github.com/waterfield/redeye.git)... _in less then 250 lines_!

## Changes from Redeye

`go-redeye` does not use redis or node.js, instead using the following substitutes:

 * Instead of redis KV, uses an in-memory hashtable
 * Instead of redis PubSub, uses Go channels
 * Instead of LUA, uses a goroutine event loop
 * Instead of Fibers for workers, uses Goroutines
 * Instead of a cache, uses shared pointers
 * Instead of one process per core, uses just one multi-core process

## Building

```sh
go build
./redeye
```

## Defining Workers

Let's say you define a _worker_ named `Fib`, which takes a single argument.
The declaration for this worker might start like:

```go
func Fib(n int) (int) {
  if n < 2 {
    return 1
  }
  return Fib(n - 1) + Fib(n - 2)
}
```

This is essentially all you need for a complete `go-redeye` worker!
However, workers are required to also return an optional error, so
our final worker code looks like [this](workers/fib.go)

```go
func Fib(n int) (int, error) {
  if n < 2 {
    return 1, nil
  }
  f1, _ := Fib(n - 1)
  f2, _ := Fib(n - 2)
  return (f1 + f2), nil
}
```

This definition is readable and looks like normal Go code. To make it
have the same magical API-feel that the Coffeescript version does, this
definition is parsed and converted using a [ruby script](script/compile.rb).
For instance, consider the following example:

```go
func Add(a, b int) (int, error) {
  return (a + b), nil
}

func PlusTwo(a int) (int, error) {
  b := 2
  return Add()
}
```

We get to call `Add()` without any arguments, because the parser knows
what the parameters of `Add` are, so it uses the local variables of `PlusTwo`
to complete the call, resulting in something like the following:

```go
func PlusTwo(a int) (int, error) {
  b := 2
  return Add(a, b)
}
```

To make `go-redeye` aware of this function, the actual compiled code looks a
little different, with extra arguments that contain some contextual information:

```go
import "fmt"

func definePlusTwo(__router *Router) {
  router.Define("PlusTwo", func(__args string) (interface{}, error) {
    var a int
    fmt.Sscanf(__args, "%v", &a)
    
    b := 2
    return Add(__router, "PlusTwo", __args, a, b)
  })
}
```

You might be wondering, what about that call to `Add`? There still needs to be
an actual Go function declaration for it. In actuality, `Add` will be generated as
a wrapper function that requests a key value from the event loop:

```go
import "fmt"

func Add(router *Router, tgtPrefix, tgtArgs string, a int, b int) (int, error) {
  args := fmt.Sprintf("%v:%v", a, b)
  value, err := router.Get("Add", args, tgtPrefix, tgtArgs)
  return value.(int), err
}
```

## Implementation

The core part of `go-redeye` is the [router](router.go).
I suggest reading the comments there, but here is a quick example of how the above
example would actually run.

 * `PlusTwo` is running in a goroutine and has local variables `a` and `b`
 * It calls `Add`, which as you saw is a wrapper that calls `router.Get`
 * `router.Get` constructs a request for the `Add:a:b` key
 * The request also contains a new unique channel
 * The request is sent to a channel on the router called `get`
 * The router's main event loop receives the `get` request
 * It checks to see if `Add:a:b` is present in the KV store
 * Since it's not, it records the unique channel from the request in a "waiters" list
 * Then it calls the internal function `work`, which spawns a new Goroutine
 * The new goroutine calls the anonymous function for `Add`, passing in the args "a:b"
 * When the `Add` routine finishes, it finds the resulting value
 * The "Add a:b" key and its return value are sent to the `finish` channel on the router
 * The event loop gets the message and records the key and value in the KV store
 * The router sees that a channel is waiting on this key in the "waiters" list
 * It takes the value from `Add` and sends it on the response channel
 * Then it removes the entry from the "waiters" list
 * The initial call to `Get`, which was blocking on its unique channel, returns the value
 * The value is now returned to the `PlusTwo` routine, which can continue

Using channels and Goroutines, the `go-redeye` process can scale transparently across CPUs
(and even machines).

## Results

This repo contains the `Fib` example. The [test](test.go) program requests `Fib(10)`,
whose definition is [here](workers/fib.go) and compiled version [here](fib.go). The test
program first prints `Fib(10)` and then prints all the KV mappings; here are the results:

```
➜ ./redeye 
Fib(10): 89 

{Fib 0} = {1 <nil>}
{Fib 9} = {55 <nil>}
{Fib 2} = {2 <nil>}
{Fib 5} = {8 <nil>}
{Fib 1} = {1 <nil>}
{Fib 6} = {13 <nil>}
{Fib 7} = {21 <nil>}
{Fib 8} = {34 <nil>}
{Fib 3} = {3 <nil>}
{Fib 4} = {5 <nil>}
{Fib 10} = {89 <nil>}
```
