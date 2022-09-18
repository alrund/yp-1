/*
Package staticlint is a tool for static analysis of Go programs.
Usage: staticlint [-flag] [package]
Run 'staticlint help' for more detail,

Registered analyzers:

	asmdecl Package asmdecl defines an Analyzer that reports mismatches between assembly files and Go declarations.
	assign Package assign defines an Analyzer that detects useless assignments.
	atomic Package atomic defines an Analyzer that checks for common mistakes using the sync/atomic package.
	atomicalign Package atomicalign defines an Analyzer that checks for non-64-bit-aligned arguments
	to sync/atomic functions.
	bools Package bools defines an Analyzer that detects common mistakes involving boolean operators.
	buildssa Package buildssa defines an Analyzer that constructs the SSA representation of an error-free package and
	returns the set of all functions within it.
	buildtag Package buildtag defines an Analyzer that checks build tags.
	cgocall Package cgocall defines an Analyzer that detects some violations of the cgo pointer passing rules.
	composite Package composite defines an Analyzer that checks for unkeyed composite literals.
	copylock Package copylock defines an Analyzer that checks for locks erroneously passed by value.
	ctrlflow Package ctrlflow is an analysis that provides a syntactic control-flow graph (CFG) for the body
	of a function.
	deepequalerrors Package deepequalerrors defines an Analyzer that checks for the use of reflect.DeepEqual
	with error values.
	errorsas The errorsas package defines an Analyzer that checks that the second argument to errors.As is a pointer
	to a type implementing error.
	findcall Package findcall defines an Analyzer that serves as a trivial example and test of the Analysis API.
	httpresponse Package httpresponse defines an Analyzer that checks for mistakes using HTTP responses.
	inspect Package inspect defines an Analyzer that provides an AST inspector
	(golang.org/x/tools/go/ast/inspector.Inspector) for the syntax trees of a package.
	loopclosure Package loopclosure defines an Analyzer that checks for references to enclosing loop variables from
	within nested functions.
	lostcancel Package lostcancel defines an Analyzer that checks for failure to call a context cancellation function.
	nilfunc Package nilfunc defines an Analyzer that checks for useless comparisons against nil.
	nilness Package nilness inspects the control-flow graph of an SSA function and reports errors such as nil pointer
	dereferences and degenerate nil pointer comparisons.
	pkgfact The pkgfact package is a demonstration and test of the package fact mechanism.
	printf Package printf defines an Analyzer that checks consistency of Printf format strings and arguments.
	shadow Package shadow defines an Analyzer that checks for shadowed variables.
	shift Package shift defines an Analyzer that checks for shifts that exceed the width of an integer.
	sortslice Package sortslice defines an Analyzer that checks for calls to sort.Slice that do not use a slice
	type as first argument.
	stdmethods Package stdmethods defines an Analyzer that checks for misspellings in the signatures of methods
	similar to well-known interfaces.
	structtag Package structtag defines an Analyzer that checks struct field tags are well formed.
	tests Package tests defines an Analyzer that checks for common mistaken usages of tests and examples.
	unmarshal The unmarshal package defines an Analyzer that checks for passing non-pointer or non-interface types
	to unmarshal and decode functions.
	unreachable Package unreachable defines an Analyzer that checks for unreachable code.
	unsafeptr Package unsafeptr defines an Analyzer that checks for invalid conversions of uintptr to unsafe.Pointer.
	unusedresult Package unusedresult defines an analyzer that checks for unused results of calls to certain
	pure functions.

	SA1000 Invalid regular expression
	SA1001 Invalid template
	SA1002 Invalid format in time.Parse
	SA1003 Unsupported argument to functions in encoding/binary
	SA1004 Suspiciously small untyped constant in time.Sleep
	SA1005 Invalid first argument to exec.Command
	SA1006 Printf with dynamic first argument and no further arguments
	SA1007 Invalid URL in net/url.Parse
	SA1008 Non-canonical key in http.Header map
	SA1010 (*regexp.Regexp).FindAll called with n == 0, which will always return zero results
	SA1011 Various methods in the 'strings' package expect valid UTF-8, but invalid input is provided
	SA1012 A nil context.Context is being passed to a function, consider using context.
	SA1013 io.Seeker.Seek is being called with the whence constant as the first argument, but it should be the second
	SA1014 Non-pointer value passed to Unmarshal or Decode
	SA1015 Using time.Tick in a way that will leak. Consider using time.NewTicker, and only use time.Tick in tests,
	commands and endless functions
	SA1016 Trapping a signal that cannot be trapped
	SA1017 Channels used with os/signal.Notify should be buffered
	SA1018 strings.Replace called with n == 0, which does nothing
	SA1019 Using a deprecated function, variable, constant or field
	SA1020 Using an invalid host:port pair with a net.Listen-related function
	SA1021 Using bytes.Equal to compare two net.IP
	SA1023 Modifying the buffer in an io.Writer implementation
	SA1024 A string cutset contains duplicate characters
	SA1025 It is not possible to use (*time.Timer).Reset's return value correctly
	SA1026 Cannot marshal channels or functions
	SA1027 Atomic access to 64-bit variable must be 64-bit aligned
	SA1028 sort.Slice can only be used on slices
	SA1029 Inappropriate key in call to context.WithValue
	SA1030 Invalid argument in call to a strconv function
	SA2000 sync.WaitGroup.Add called inside the goroutine, leading to a race condition
	SA2001 Empty critical section, did you mean to defer the unlock?
	SA2002 Called testing.T.FailNow or SkipNow in a goroutine, which isn't allowed
	SA2003 Deferred Lock right after locking, likely meant to defer Unlock instead
	SA3000 TestMain doesn't call os.Exit, hiding test failures
	SA3001 Assigning to b.N in benchmarks distorts the results
	SA4000 Binary operator has identical expressions on both sides
	SA4001 &*x gets simplified to x, it does not copy x
	SA4003 Comparing unsigned values against negative values is pointless
	SA4004 The loop exits unconditionally after one iteration
	SA4005 Field assignment that will never be observed. Did you mean to use a pointer receiver?
	SA4006 A value assigned to a variable is never read before being overwritten. Forgotten error check or dead code?
	SA4008 The variable in the loop condition never changes, are you incrementing the wrong variable?
	SA4009 A function argument is overwritten before its first use
	SA4010 The result of append will never be observed anywhere
	SA4011 Break statement with no effect. Did you mean to break out of an outer loop?
	SA4012 Comparing a value against NaN even though no value is equal to NaN
	SA4013 Negating a boolean twice (!!b) is the same as writing b. This is either redundant, or a typo.
	SA4014 An if/else if chain has repeated conditions and no side-effects; if the condition didn't match the
	first time, it won't match the second time, either
	SA4015 Calling functions like math.Ceil on floats converted from integers doesn't do anything useful
	SA4016 Certain bitwise operations, such as x ^ 0, do not do anything useful
	SA4018 Self-assignment of variables
	SA4019 Multiple, identical build constraints in the same file
	SA4020 Unreachable case clause in a type switch
	SA4021 'x = append(y)' is equivalent to 'x = y'
	SA4022 Comparing the address of a variable against nil
	SA4023 Impossible comparison of interface value with untyped nil
	SA4024 Checking for impossible return value from a builtin function
	SA4025 Integer division of literals that results in zero
	SA4026 Go constants cannot express negative zero
	SA4027 (*net/url.URL).Query returns a copy, modifying it doesn't change the URL
	SA4028 x % 1 is always zero
	SA4029 Ineffective attempt at sorting slice
	SA4030 Ineffective attempt at generating random number
	SA4031 Checking never-nil value against nil
	SA5000 Assignment to nil map
	SA5001 Deferring Close before checking for a possible error
	SA5002 The empty for loop ('for {}') spins and can block the scheduler
	SA5003 Defers in infinite loops will never execute
	SA5004 'for { select { ...' with an empty default branch spins
	SA5005 The finalizer references the finalized object, preventing garbage collection
	SA5007 Infinite recursive call
	SA5008 Invalid struct tag
	SA5009 Invalid Printf call
	SA5010 Impossible type assertion
	SA5011 Possible nil pointer dereference
	SA5012 Passing odd-sized slice to function expecting even size
	SA6000 Using regexp.Match or related in a loop, should use regexp.Compile
	SA6001 Missing an optimization opportunity when indexing maps by byte slices
	SA6002 Storing non-pointer values in sync.Pool allocates memory
	SA6003 Converting a string to a slice of runes before ranging over it
	SA6005 Inefficient string comparison with strings.ToLower or strings.ToUpper
	SA9001 Defers in range loops may not run when you expect them to
	SA9002 Using a non-octal os.FileMode that looks like it was meant to be in octal.
	SA9003 Empty body in an if or else branch
	SA9004 Only the first constant has an explicit type
	SA9005 Trying to marshal a struct with no public fields nor custom marshaling
	SA9006 Dubious bit shifting of a fixed size integer value
	SA9007 Deleting a directory that shouldn't be deleted
	SA9008 else branch of a type assertion is probably not reading the right value

	ST1005 Incorrectly formatted error string

	osexit os.Exit presence for the main function in the main package
*/
package main

import (
	"github.com/alrund/yp-1/cmd/staticlint/osexitanalyzer"
	"github.com/charithe/durationcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	mychecks := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		findcall.Analyzer,
		httpresponse.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}

	for name, v := range staticcheck.Analyzers {
		if name == "SA4017" { // SA4017 panic
			continue
		}
		mychecks = append(mychecks, v)
	}

	mychecks = append(mychecks, stylecheck.Analyzers["ST1005"])

	mychecks = append(mychecks, bodyclose.Analyzer)
	mychecks = append(mychecks, durationcheck.Analyzer)

	mychecks = append(mychecks, osexitanalyzer.OsExitAnalyzer)

	multichecker.Main(mychecks...)
}
