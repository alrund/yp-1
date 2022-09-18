package main

import "os"

func main() {
	os.Exit(1) // want "os.exit detected"
}

func foo() {
	os.Exit(1)
}
