package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/bluebrown/mdhero"
)

func main() {
	var mdflags mdhero.Flags
	flag.Var(bit(&mdflags, mdhero.DEBUG), "debug", "print debug messages")
	flag.Var(bit(&mdflags, mdhero.HTML), "html", "enable HTML output")
	flag.Var(bit(&mdflags, mdhero.BROWSER), "browser", "open in browser")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: md [-debug] [-html] [-browser] <source> [<target>]")
	}

	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(64)
	}

	md := mdhero.New(flag.Args()[0], mdhero.WithFlags(mdflags))

	if len(flag.Args()) > 1 {
		md.Target = flag.Args()[1]
	}

	if err := md.Run(); err != nil {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(65)
	}
}

func bit(field *mdhero.Flags, mask mdhero.Flags) *bitFlag {
	return &bitFlag{Flags: field, Mask: mask}
}

type bitFlag struct {
	Flags *mdhero.Flags
	Mask  mdhero.Flags
}

func (f *bitFlag) String() string {
	if f.Flags == nil {
		return "false"
	}
	return strconv.FormatBool((*f.Flags & f.Mask) != 0)
}

func (f *bitFlag) IsBoolFlag() bool {
	return true
}

func (f *bitFlag) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	if v {
		*f.Flags |= f.Mask
	} else {
		*f.Flags &^= f.Mask
	}
	return nil
}
