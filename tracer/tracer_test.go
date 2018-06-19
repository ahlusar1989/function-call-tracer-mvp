package tracer

import (
	"bytes"
	"log"

	"github.com/stretchr/testify/assert"
	"testing"
)

// Define a custom logger which is just a string buffer
// so we can validate that we are building the appropriate
// trace messages
var TestBuffer bytes.Buffer
var BufLogger = log.New(&TestBuffer, "", 0)

func GetTestBuffer() string {
	return TestBuffer.String()
}

func ResetTestBuffer() {
	// This prepends a newline in front of the buffer after a reset
	// since our validation logic below opens the on the previous line
	// to make the validation logic easier to read, and copy in and out
	TestBuffer.Reset()
	BufLogger.Printf("\n")
}

func TestBasicUsage(test *testing.T) {
	ResetTestBuffer()
	Exit, Enter := NewFunctionTracer(&TracerConfiguration{CustomLogger: BufLogger})

	second := func() {
		defer Exit(Enter("SECOND"))
	}
	first := func() {
		defer Exit(Enter("FIRST"))
		second()
	}
	first()

	assert.Equal(test, GetTestBuffer(), `
[ 0]Entering Function: FIRST
[ 1]  Entering Function: SECOND
[ 1]  Exiting Function:  SECOND
[ 0]Exiting Function:  FIRST
`)
}

func TestDisableTracing(test *testing.T) {
	ResetTestBuffer()
	Exit, Enter := NewFunctionTracer(&TracerConfiguration{CustomLogger: BufLogger, DisableTracing: true})

	second := func() {
		defer Exit(Enter("SECOND"))
	}
	first := func() {
		defer Exit(Enter("FIRST"))
		second()
	}
	first()

	assert.Equal(test, GetTestBuffer(), "\n")
}

func TestCustomEnterExit(test *testing.T) {
	ResetTestBuffer()
	Exit, Enter := NewFunctionTracer(&TracerConfiguration{CustomLogger: BufLogger, EnterMessage: "enter: ", ExitMessage: "exit:  "})

	second := func() {
		defer Exit(Enter("SECOND"))
	}
	first := func() {
		defer Exit(Enter("FIRST"))
		second()
	}
	first()

	assert.Equal(test, GetTestBuffer(), `
[ 0]enter: FIRST
[ 1]  enter: SECOND
[ 1]  exit:  SECOND
[ 0]exit:  FIRST
`)
}

func TestDisableNesting(test *testing.T) {
	ResetTestBuffer()
	Exit, Enter := NewFunctionTracer(&TracerConfiguration{CustomLogger: BufLogger, DisableNesting: true})

	second := func() {
		defer Exit(Enter("SECOND"))
	}
	first := func() {
		defer Exit(Enter("FIRST"))
		second()
	}
	first()

	assert.Equal(test, GetTestBuffer(), `
[ 0]Entering Function: FIRST
[ 1]Entering Function: SECOND
[ 1]Exiting Function:  SECOND
[ 0]Exiting Function:  FIRST
`)
}

func TestCustomSpacesPerIndent(test *testing.T) {
	ResetTestBuffer()
	Exit, Enter := NewFunctionTracer(&TracerConfiguration{CustomLogger: BufLogger, SpacesPerIndent: 3})

	second := func() {
		defer Exit(Enter("SECOND"))
	}
	first := func() {
		defer Exit(Enter("FIRST"))
		second()
	}
	first()

	assert.Equal(test, GetTestBuffer(), `
[ 0]Entering Function: FIRST
[ 1]   Entering Function: SECOND
[ 1]   Exiting Function:  SECOND
[ 0]Exiting Function:  FIRST
`)
}

func TestDisableDepthValue(test *testing.T) {
	ResetTestBuffer()
	Exit, Enter := NewFunctionTracer(&TracerConfiguration{CustomLogger: BufLogger, DisableDepthValue: true})

	second := func() {
		defer Exit(Enter("SECOND"))
	}
	first := func() {
		defer Exit(Enter("FIRST"))
		second()
	}
	first()

	assert.Equal(test, GetTestBuffer(), `
Entering Function: FIRST
  Entering Function: SECOND
  Exiting Function:  SECOND
Exiting Function:  FIRST
`)
}

// Helper function - part of "TestUnspecifiedFunctionName"
func foobar() {
	Exit, Enter := NewFunctionTracer(&TracerConfiguration{CustomLogger: BufLogger})
	defer Exit(Enter())
}
func TestUnspecifiedFunctionName(test *testing.T) {
	ResetTestBuffer()

	// Call another named function
	foobar()

	assert.Equal(test, GetTestBuffer(), `
[ 0]Entering Function: function-call-trace/tracer.foobar
[ 0]Exiting Function:  function-call-trace/tracer.foobar
`)
}

// Negative tests
func TestMoreExitsThanEntersMustPanic(test *testing.T) {
	Exit, _ := NewFunctionTracer(&TracerConfiguration{CustomLogger: BufLogger})
	assert.Panics(test, func() {
		Exit("")
	}, "Calling exit without enter should panic")
}

// Examples
func ExampleNew_noOptions() {
	Exit, Enter := NewFunctionTracer(nil)

	second := func() {
		defer Exit(Enter("SECOND"))
	}
	first := func() {
		defer Exit(Enter("FIRST"))
		second()
	}
	first()

	// Output:
	// [ 0]Entering Function: FIRST
	// [ 1]  Entering Function: SECOND
	// [ 1]  Exiting Function:  SECOND
	// [ 0]Exiting Function:  FIRST
}

func ExampleNew_customMessage() {
	Exit, Enter := NewFunctionTracer(&TracerConfiguration{EnterMessage: "en - ", ExitMessage: "ex - "})

	second := func() {
		defer Exit(Enter("SECOND"))
	}
	first := func() {
		defer Exit(Enter("FIRST"))
		second()
	}
	first()

	// Output:
	// [ 0]en - FIRST
	// [ 1]  en - SECOND
	// [ 1]  ex - SECOND
	// [ 0]ex - FIRST
}

func ExampleNew_changeIndentLevel() {
	Exit, Enter := NewFunctionTracer(&TracerConfiguration{SpacesPerIndent: 1})

	second := func() {
		defer Exit(Enter("SECOND"))
	}
	first := func() {
		defer Exit(Enter("FIRST"))
		second()
	}
	first()

	// Output:
	// [ 0]Entering Function: FIRST
	// [ 1] Entering Function: SECOND
	// [ 1] Exiting Function:  SECOND
	// [ 0]Exiting Function:  FIRST
}