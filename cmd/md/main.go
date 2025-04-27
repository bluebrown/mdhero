package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/bluebrown/mdhero"
)

func main() {
	var (
		mdflags = mdhero.Flags(0)
	)

	flag.Var(&BitFlag[mdhero.Flags]{Field: &mdflags, Mask: mdhero.DEBUG}, "debug", "print debug messages")
	flag.Var(&BitFlag[mdhero.Flags]{Field: &mdflags, Mask: mdhero.HTML}, "html", "enable HTML output")

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

type BitFlag[T ~uint8] struct {
	Field *T
	Mask  T
}

func (f *BitFlag[T]) String() string {
	if f.Field == nil {
		return "nil"
	}
	return strconv.FormatUint(uint64(*f.Field), 10)
}

func (f *BitFlag[T]) IsBoolFlag() bool {
	return true
}

func (f *BitFlag[T]) Set(s string) error {
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
