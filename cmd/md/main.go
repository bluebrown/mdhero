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

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: md [-debug] [-html] <source> [<target>]")
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
	return &bitFlag{Field: field, Mask: mask}
}

type bitFlag struct {
	Field *mdhero.Flags
	Mask  mdhero.Flags
}

func (f *bitFlag) String() string {
	if f.Field == nil {
		return "false"
	}
	return strconv.FormatBool((*f.Field & f.Mask) != 0)
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
		*f.Field |= f.Mask
	} else {
		*f.Field &^= f.Mask
	}
	return nil
}
