package util

import "fmt"

// ReturnErrOnPanic catches panics, expects them to be of type error, then stores it in the pointer as recovery
func ReturnErrOnPanic(perr *error) func() {
	return func() {
		if r := recover(); r != nil {
			*perr = r.(error)
		}
	}
}

// Ensure checks for an error. If there is one it panics, if no error it prints the message (given it is not empty)
// `message` should describe what is ensured by the call and verification of `err` being `nil`
func Ensure(err error, message string) {
	if err != nil {
		panic(err)
	}
	if message != "" {
		fmt.Println(message)
	}
}

// PanicOnErr panics if the passed error != nil
func PanicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
