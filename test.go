package main

import "fmt"

// function to print hello
func printHello() {
	fmt.Println("Hello from printHello")
}

func main() {
	//inline goroutine. Define a function inline and then call it.
	go func() { fmt.Println("Hello inline") }()
	//call a function as goroutine
	go printHello()
	fmt.Println("Hello from main")
}
