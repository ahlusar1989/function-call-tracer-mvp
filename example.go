package main

import (
	"./tracer"
)

// Setup global enter exit trace functions (default options)
// See tests in tracer package for more advanced usage
var Exit, Enter = tracer.NewFunctionTracer(nil)
// Or...
// var Exit, Enter = tracer.NewFunctionTracer(nil)

func foo(i int) {
    // #FN will get replaced with the function's name
    defer Exit(Enter("#FN(%d)", i))
    if i != 0 {
        foo(i - 1)
    }
}

func Foo() {
    defer Exit(Enter("#FN is awesome %d %s", 3, "four"))
}


// The Enter() receives a variadic list interfaces: ...interface{}. 
// This allows us to pass in a variable number of types.
 // However, the first of such is expected to be a format string, 
// otherwise the function just logs the function's name. If a 
// format string is specified with a $#FN token, then said token is 
// replaced for the actual function's name.

func main() {
    defer Exit(Enter())
    foo(2)
    Foo()
}