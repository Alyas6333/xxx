// xxx is a debuging library for golang.
package xxx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"

	"gopkg.in/mgo.v2"

	"github.com/davecgh/go-spew/spew"
	"github.com/juju/juju/mongo"
)

// Print prints v along with filename and line number of call.
func Print(v interface{}, message ...string) {
	called(2, fmt.Sprintf("%s \n%#v\n", m(message), v))
}

// Printf prints format along with filename and line number of call.
func Printf(format string, vars ...interface{}) {
	called(2, fmt.Sprintf(format, vars...))
}

// Dump prints all information on x.
func Dump(x ...interface{}) {
	if off {
		return
	}
	called(2, "spew")
	var a []interface{}
	a = append(a, x...)
	spew.Dump(a)
}

// Stack prints call stack to this point.
func Stack() {
	if off {
		return
	}
	debug.PrintStack()
}

// DumpColl dumps the contents of a mongo database collection.
func DumpColl(db *mgo.Database, collName string) {
	if off {
		return
	}
	coll, closer := mongo.CollectionFromName(db, collName)
	defer closer()

	var results []interface{}
	err := coll.Find(nil).All(&results)
	if err != nil {
		panic(err)
	}
	called(2, collName)
	if len(results) == 0 {
		fmt.Println("no results found")
		return
	}
	for i, r := range results {
		output, err := json.MarshalIndent(r, "", " ")
		if err != nil {
			fmt.Print(err.Error())
		}
		fmt.Printf("Doc %d (%T): %s\n", i, r, output)
		fmt.Println("")
	}
	fmt.Println("")
}

// CaptureStdOutAndErr captures all std out/err until the returned func is
// called. That func returns any captured output as a string. This is useful
// for capturing remote output (e.g. on a server) and writing it to a file or
// piping it to a log etc.
func CaptureStdOutAndErr() func() string {
	old := os.Stdout // keep backup of the real stdout
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w

	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	outErrC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rErr)
		outErrC <- buf.String()
	}()

	return func() string {
		w.Close()
		wErr.Close()
		os.Stdout = old    // restoring the real stdout
		os.Stderr = oldErr // restoring the real stdout
		return <-outC + <-outErrC
	}
}

func Called(message ...string) {
	called(2, m(message))
}

func CalledX(x int, message ...string) {
	c := 2 + x
	called(c, m(message))
}

var off bool

// Off stops any output being printed from this package.
func Off() {
	off = true
}

// On starts any output being printed from this package. When you have code
// peppered with debugging output, use On and Off to print just the debug
// outputs you are interested in.
func On() {
	off = false
}

func called(scope int, message string) {
	if off {
		return
	}
	if _, file, line, ok := runtime.Caller(scope); ok {
		fmt.Printf("%s:%d: %s \n", file, line, message)
	}
}

func m(ms []string) string {
	var m string
	if ms != nil {
		m = ms[0]
	}
	return m
}
