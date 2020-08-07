package flags

import (
	"flag"
	"log"
	"os"
)

func NewFlagSet(name string) *FlagSet {

	return &FlagSet{FlagSet: flag.NewFlagSet(name, flag.ExitOnError)}
}

type FlagSet struct {
	*flag.FlagSet
}

func (fs *FlagSet) BoolFlag(name string, value bool, usage string) *BoolFlag {

	f := fs.FlagSet.Bool(name, value, usage)

	strFlag := &BoolFlag{Name: name, Default: value, ptr: f, checker: &checker{name: name, visitFunc: fs.Visit}}

	return strFlag
}

func (fs *FlagSet) StringFlag(name string, value string, usage string) *StringFlag {

	strFlag := &StringFlag{Name: name, Default: value, ptr: new(string), checker: &checker{name: name, visitFunc: fs.Visit}}

	fs.FlagSet.StringVar(strFlag.ptr, name, value, usage)

	return strFlag
}

func (fs *FlagSet) StringArrFlag(name, value, usage string) *StringArrFlag {
	saf := &StringArrFlag{Name: name, Default: value, checker:&checker{name: name, visitFunc: fs.Visit}}

	fs.FlagSet.Var(saf, name, usage)

	return saf
}

func (fs *FlagSet) IntFlag(name string, value int, usage string) *IntFlag {

	f := &IntFlag{Name: name, Default: value, ptr: new(int), checker: &checker{name: name, visitFunc: fs.Visit}}

	fs.FlagSet.IntVar(f.ptr, name, value, usage)

	return f
}

func (fs *FlagSet) ParseArgs() *CommandArgs {

	cmdArgs := os.Args[1:]
	if err := fs.FlagSet.Parse(cmdArgs); err != nil {
		log.Fatal(err)
	}

	return &CommandArgs{args: fs.Args()}
}
