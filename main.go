package main

import (
	"syscall/js"
)

func main() {
	window := js.Global()
	document := window.Get("document")
	body := document.Get("body")
	p := document.Call("createElement", "p")
	p.Set("textContent", "Hello, World! (from Go WASM)")
	body.Call("appendChild", p)
	<-make(chan bool)
}
