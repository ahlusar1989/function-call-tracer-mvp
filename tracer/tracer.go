package tracer


import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"reflect"
	"runtime"
)

// Define global regexes for extracting function names
var regexPredecessor = regexp.MustCompile(`^.*\.(.*)#`)
var regexDetectFunc = regexp.MustCompile(`\#FN`)

type TracerConfiguration struct {

	// Setting "DisableTracing" to "true" will cause this tracer to return
	// no-op'd functions for both exit() and enter(). The default value
	// for this is "false" which enables tracing.
	DisableTracing bool

	// Setting the "CustomLogger" to nil will cause this tracer to log to
	// os.Stdout. Otherwise, this is a pointer to an object as returned
	// from `log.New(...)`.
	CustomLogger *log.Logger

	// Setting "DisableDepthValue" to "true" will cause this tracer to not
	// prepend the printed function's depth to enter() and exit() messages.
	// The default value is "false", which logs the depth value.
	DisableDepthValue bool

	// Setting "DisableNesting" to "true" will cause this tracer to not indent
	// any messages from nested functions. The default value is "false"
	// which enables nesting by prepending "SpacesPerIndent" number of
	// spaces per level nested.
	DisableNesting  bool
	SpacesPerIndent int `default:"2"`

	// Setting "EnterMessage" or "ExitMessage" will override the default
	// value of "Entering Function " and "Exiting Function:  " respectively.
	EnterMessage string `default:"Entering Function: "`
	ExitMessage  string `default:"Exiting Function:  "`

	// Private member, used to keep track of how many levels of nesting
	// the current trace functions have navigated.
	currentDepth int
}


// Main entry-point for the lib. Calling NewFunctionTracer with nil will
// result in the default options being used.
func NewFunctionTracer(opts *TracerConfiguration) (func(string), func(...interface{}) string) {
	var options TracerConfiguration
	if opts != nil {
		options = *opts
	}

	// If tracing is not enabled, just return no-op functions
	if options.DisableTracing {
		return func(string) {}, func(...interface{}) string { return "" }
	}

	// Revert to stdout if no logger is defined
	if options.CustomLogger == nil {
		options.CustomLogger = log.New(os.Stdout, "", 0)
	}

	// Use reflect to extract default values for the
	// Entering and Exiting messages (if they are not set)
	reflectedTypes := reflect.TypeOf(options)
	if options.EnterMessage == "" {
		field, _ := reflectedTypes.FieldByName("EnterMessage")
		options.EnterMessage = field.Tag.Get("default")
	}
	if options.ExitMessage == "" {
		field, _ := reflectedTypes.FieldByName("ExitMessage")
		options.ExitMessage = field.Tag.Get("default")
	}

	// If nesting is enabled, and the spaces are not specified,
	// use the "default" value
	if options.DisableNesting {
		options.SpacesPerIndent = 0
	} else if options.SpacesPerIndent == 0 {
		field, _ := reflectedTypes.FieldByName("SpacesPerIndent")
		options.SpacesPerIndent, _ = strconv.Atoi(field.Tag.Get("default"))
	}

	//
	// Define functions we will use and return to the caller
	//
	spaceOpt := func() string {
		spaces := strings.Repeat(" ", options.currentDepth*options.SpacesPerIndent)
		if !options.DisableDepthValue {
			return fmt.Sprintf("[%2d]%s", options.currentDepth, spaces)
		}
		return spaces
	}

	// Increment function to increase the current depth value
	incrementDepthCount := func() {
		options.currentDepth += 1
	}

	// Decrement function to decrement the current depth value
	//  + panics if current depth value is < 0
	decrementCallDepth := func() {
		options.currentDepth -= 1
		if options.currentDepth < 0 {
			panic("Depth is negative! This should never happen")
		}
	}

	// Enter function, invoked on function entry
	enterFunc := func(args ...interface{}) string {
		defer incrementDepthCount()

		// Figure out the name of the caller and use that
		fnName := "<unknown>"
		pc, _, _, ok := runtime.Caller(1)
		if ok {
			// match on text for the fist submatch
			fnName = regexPredecessor.ReplaceAllString(runtime.FuncForPC(pc).Name(), "$1")
		}

		traceMessage := fnName
		if len(args) > 0 {
			if fmtStr, ok := args[0].(string); ok {
				// We have a string leading args, assume its to be formatted
				traceMessage = fmt.Sprintf(fmtStr, args[1:]...)
			}
		}

		// "#FN" will be replaced by the name of the function (if present)
		traceMessage = regexDetectFunc.ReplaceAllString(traceMessage, fnName)

		options.CustomLogger.Printf("%s%s%s\n", spaceOpt(), options.EnterMessage, traceMessage)
		return traceMessage
	}

	// Exit function, invoked on function exit (usually deferred)
	exitFunc := func(s string) {
		decrementCallDepth()
		options.CustomLogger.Printf("%s%s%s\n", spaceOpt(), options.ExitMessage, s)
	}

	return exitFunc, enterFunc
}
