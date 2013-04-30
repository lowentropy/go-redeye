package main

import (
	"fmt"
)

// worker defines how the body of a worker function should behave. It
// should take a router object and argument string, and return a generic
// value and optional error.
type worker func(args string) (interface{}, error)

// key defines a unique key within the system. It consists of a prefix
// that uniquely identifies a worker function, as well as a string of
// combined arguments.
type key struct {
	prefix, args string
}

// value defines the result of a worker function, stored as the result
// for a particular key. It consists of a generic data object for the return
// value, plus an optional error.
type value struct {
	data interface{}
	err  error
}

// internal redeye errors use this
type redeyeError struct {
	key     key
	message string
}

// Router is the object that controls message routing and memoization database
// access for redeye workers. Its API communicates with a goroutine event loop
// using channels.
type Router struct {
	data    map[key]value        // map from keys to values
	waiters map[key][]chan value // list of channels waiting for each key
	workers map[string]worker    // worker bodies defined per prefix
	active  map[key]bool         // whether a worker exists for the key
	sources map[key][]key        // map of sources of each key
	get     chan getCmd          // send get requests here
	finish  chan finishCmd       // send finished keys here
	quit    chan bool            // send on here to quit the event loop
}

// getCmd is a struct sent to the router's `get` channel to request
// a lookup for a key. It includes the key to be looked up, as well
// as a channel on which the result should be sent. If the key is
// not already computed, a new worker will be created for it.
type getCmd struct {
	key key
	ch  chan value
}

// finishCmd is a struct sent to the router's `finish` channel to
// indicate that work on a particular key is done. It includes the
// key that was worked on, as well as the resulting value for that key.
type finishCmd struct {
	key   key
	value value
}

// New constructs a new router; it is not automatically started.
func New() *Router {
	router := new(Router)
	router.data = make(map[key]value)
	router.waiters = make(map[key][]chan value)
	router.workers = make(map[string]worker)
	router.active = make(map[key]bool)
	router.sources = make(map[key][]key)
	router.get = make(chan getCmd)
	router.finish = make(chan finishCmd)
	router.quit = make(chan bool)
	return router
}

// Get is the API to request a key from the router. It takes the prefix
// and arguments of the key to be looked up, and returns a generic value
// and optional error. The request will block until the value is available.
// It does this by creating a response channel and sending it in a new getCommand
// to the router; it then listens on the channel for the return value.
func (router *Router) Get(srcPrefix, srcArgs, tgtPrefix, tgtArgs string) (interface{}, error) {
	src := key{srcPrefix, srcArgs}
	tgt := key{tgtPrefix, tgtArgs}
	if err := addLink(router, src, tgt); err != nil {
		return nil, err
	}
	ch := make(chan value)
	router.get <- getCmd{src, ch}
	value := <-ch
	return value.data, value.err
}

// Define will define the worker body for the given prefix
func (router *Router) Define(prefix string, body worker) {
	router.workers[prefix] = body
}

// Start will start a goroutine that
func (router *Router) Start() {
	go route(router)
}

// Causes the main event loop to terminate, which means any
// subsequent Get() commands, or any finishing worker processes,
// will block and hang the process. Please make sure all work
// is actually done before calling this.
func (router *Router) Quit() {
	router.quit <- true
}

// route is the main event loop. It lists for getCmd and finishCmd
// events and calls the get and finish methods. It is organized this
// way to create a channel-based lock around the two important
// maps, data and waiters.
func route(router *Router) {
loop:
	for {
		select {
		case cmd := <-router.get:
			get(router, cmd)
		case cmd := <-router.finish:
			finish(router, cmd)
		case <-router.quit:
			break loop
		}
	}
}

// get tries to satisfy a getCmd. If the requested key is already complete,
// it immediately sends the associated value back on the response channel.
// Otherwise, it records the response channel as a waiter for the given key.
// Also, if the key has not been requested before, it will call the work method
// to start a goroutine to satisfy the key.
func get(router *Router, cmd getCmd) {
	value, ok := router.data[cmd.key]
	if ok {
		cmd.ch <- value
		return
	}
	wait(router, cmd)
	if _, ok := router.active[cmd.key]; !ok {
		work(router, cmd.key)
	}
}

// finish sets the value of the given key. If any channels are waiting on
// the key to be done, it sends the result to those channels, then deletes
// them from the waiters map.
func finish(router *Router, cmd finishCmd) {
	router.data[cmd.key] = cmd.value
	waiters, ok := router.waiters[cmd.key]
	if !ok {
		return
	}
	delete(router.waiters, cmd.key)
	for _, ch := range waiters {
		ch <- cmd.value
	}
}

// wait records a channel as wanting to wait on the given
// key to complete.
func wait(router *Router, cmd getCmd) {
	if _, ok := router.waiters[cmd.key]; !ok {
		router.waiters[cmd.key] = []chan value{cmd.ch}
	} else {
		router.waiters[cmd.key] = append(router.waiters[cmd.key], cmd.ch)
	}
}

// work will look up the defined worker body for the given prefix.
// If none is found, it records the value of the key as being an error.
// Otherwise, it calls the body function in a new gorouting. On completion
// of the body function, the result is sent to the router's finish channel.
func work(router *Router, key key) {
	router.active[key] = true
	body, ok := router.workers[key.prefix]
	if !ok {
		err := &redeyeError{key, "No runner for prefix"}
		go func() {
			router.finish <- finishCmd{key, value{nil, err}}
		}()
	} else {
		go func() {
			data, err := body(key.args)
			router.finish <- finishCmd{key, value{data, err}}
		}()
	}
}

// addLink first calls checkCycle and returns an error if any.
// then, it adds a mapping between src and tgt in the router's
// sources map
func addLink(router *Router, src, tgt key) error {
	if err := checkCycle(router, src, tgt); err != nil {
		return err
	}
	if _, ok := router.sources[tgt]; !ok {
		router.sources[tgt] = []key{src}
	} else {
		router.sources[tgt] = append(router.sources[tgt], src)
	}
	return nil
}

// checCycle determines if there's a cycle. This will be the case
// when src == tgt; do a depth-first search across the
// sources map, ignoring keys where a value exists or the
// key is not active yet. If there is a cycle, will return an
// error, otherwise returns nil.
func checkCycle(router *Router, src, tgt key) error {
	// if src == target, we found a cycle
	if src == tgt {
		return &redeyeError{tgt, "Cycle!"}
	}
	// skip keys with values, since they can't have caused cycles
	if _, ok := router.data[src]; ok {
		return nil
	}
	// skip inactive keys, because we don't know their sources
	if !router.active[src] {
		return nil
	}
	// if we don't have sources for the key, quit
	sources, ok := router.sources[src]
	if !ok {
		return nil
	}
	// recursively call for each dependency; return the first error
	for _, dep := range sources {
		if err := checkCycle(router, dep, tgt); err != nil {
			return err
		}
	}
	// no cycle was found
	return nil
}

// Error function for the redeyeError struct
func (err *redeyeError) Error() string {
	return fmt.Sprintf("In %v:%v: %v", err.key.prefix, err.key.args, err.message)
}
