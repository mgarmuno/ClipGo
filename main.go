package main

import "github.com/goki/gi/gi"

func main() {
	win := gi.NewMainWindow("demo", "Demo", 600, 400)
	win.StartEventLoop()
}
