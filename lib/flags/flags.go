package flags

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
)

func StringArr(name, value, usage string) *StringArrFlag {
	saf := &StringArrFlag{Name: name, Default: value, checker:&checker{name: name}}

	flag.Var(saf, name, usage)

	return saf
}

// A string flag that can be passed multiple times
type StringArrFlag struct {
	Name    string
	Default string

	vals []string
	*checker
}

func (f *StringArrFlag) String() string {

	return f.Default
}

func (f *StringArrFlag) Set(value string) error {
	f.vals = append(f.vals, value)
	return nil
}

func (f *StringArrFlag) Values() []string {
	return f.vals
}

// The main argument of a cli tool
type Arg struct {
	val string
}

func (a *Arg) Val() string {
	return a.val
}

// Return true if argument is file, and return the bytes of the file
// Otherwise return false and nil
func (a *Arg) IsFile() (bool, []byte) {

	if match := regexp.MustCompile(`(\.\/|\/)?(\w\/?\.?)+`).MatchString(a.Val()); !match {
		return false, nil
	}

	bytes, err := ioutil.ReadFile(a.Val())
	if err != nil {
		return false, nil
	}

	return true, bytes
}

// Return true if argument is a parsable url, and return the parsed url
// Otherwise return false and nil
func (a *Arg) IsURL() (bool, *url.URL) {

	if match := regexp.MustCompile(`((https?://)?.*:(\d+)?).*`).MatchString(a.Val()); !match {
		return false, nil
	}

	if regexp.MustCompile(`https?://`).MatchString(a.Val()) {
		return parseUrl(a.Val())
	}

	return parseUrl(fmt.Sprintf(`http://%s`, a.Val()))
}




// Flag representing string
type StringFlag struct {
	Name    string
	Default string
	ptr     *string
	*checker
}

func (f *StringFlag) Get() string {
	return *f.ptr
}

// Flag representing a boolean
type BoolFlag struct {
	Name    string
	Default bool
	ptr     *bool
	*checker
}

func (f *BoolFlag) Get() bool {
	return *f.ptr
}

// Flag representing a int
type IntFlag struct {
	Name    string
	Default int
	ptr     *int
	*checker
}

func (f *IntFlag) Get() int {
	return *f.ptr
}

type CommandArgs struct {
	args []string
}

func (a *CommandArgs) HasSize(size int) bool {
	return len(a.args) == size
}

func (a *CommandArgs) IsEmpty() bool {
	return len(a.args) == 0
}

func (a *CommandArgs) First() *Arg {
	if a.IsEmpty() {
		panic("no program arguments provided")
	}

	return &Arg{val: a.args[0]}
}

// Create a new StringFlag
func String(name, value, usage string) *StringFlag {
	f := flag.String(name, value, usage)

	return &StringFlag{Name: name, Default: value, ptr: f, checker:&checker{name: name}}
}

// Create a new BooleanFlag
func Bool(name string, value bool, usage string) *BoolFlag {
	f := flag.Bool(name, value, usage)

	return &BoolFlag{Name: name, Default: value, ptr: f, checker:&checker{name: name}}
}

// Create a new IntFlag
func Int(name string, value int, usage string) *IntFlag {
	f := flag.Int(name, value, usage)

	return &IntFlag{Name: name, Default: value, ptr: f, checker: &checker{name: name}}
}

// Retrieve the program arguments
func Args() *CommandArgs {
	return &CommandArgs{args: flag.Args()}
}

// Parse the program arguments
func Parse() {
	flag.Parse()
}

type checker struct {
	visitFunc func(fn func(*flag.Flag))
	name string
}

// Check if a given flag was set
func (c *checker) IsSet() bool {
	if c.visitFunc == nil {
		return isPassed(c.name, flag.Visit)
	}

	return isPassed(c.name, c.visitFunc)
}

// Check if a flag with given name was passed as program arguments
func isPassed(flagName string, visitFunc func(fn func(*flag.Flag))) bool {
	var set = false;
	visitFunc(func(visitedFlag *flag.Flag) {
		if !set {
			set = visitedFlag.Name == flagName
		}
	})

	return set
}

func parseUrl(val string) (bool, *url.URL) {
	parsedUrl, err := url.Parse(val)

	if err != nil {
		return false, nil
	}

	return true, parsedUrl
}