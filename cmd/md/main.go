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
		fmt.Fprintln(os.Stderr, "usage: md [-debug] [-html] <source>")
	}

	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(64)
	}

	opts := []mdhero.Option{
		mdhero.WithFlags(mdflags),
	}

	if err := mdhero.Run(flag.Args()[0], opts...); err != nil {
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
		return "nil"
	}
	return strconv.FormatUint(uint64(*f.Field), 10)
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
