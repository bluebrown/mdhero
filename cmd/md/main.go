package main

import (
	"flag"
	"log/slog"
	"os"
	"strconv"

	"github.com/bluebrown/mdhero"
)

func main() {
	var (
		logLevel = slog.LevelInfo
		mdflags  = mdhero.Flags(0)
	)

	flag.Var(&BitFlag[mdhero.Flags]{Field: &mdflags, Mask: mdhero.DEBUG}, "debug", "print debug messages")

	flag.Parse()

	if mdflags&mdhero.DEBUG != 0 {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))

	if len(flag.Args()) < 1 {
		slog.Error("no input file specified")
		os.Exit(1)
	}

	opts := []mdhero.Option{
		mdhero.WithFlags(mdflags),
	}

	if err := mdhero.Run(flag.Args()[0], opts...); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
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
