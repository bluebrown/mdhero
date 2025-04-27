package mdhero

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/charmbracelet/glamour"
)

func Run(source string, options ...Option) error {
	return New(source, options...).Run()
}

type Engine struct {
	Title  string
	Source string
	Target string
	Flags  Flags
	Chroma string
}

func New(source string, options ...Option) *Engine {
	m := &Engine{
		Source: source,
		Chroma: "nordic",
	}

	for _, opt := range options {
		opt(m)
	}

	return m
}

type Flags uint8

const (
	DEBUG Flags = 1 << iota
)

type Option func(*Engine)

func WithTitle(title string) Option {
	return func(m *Engine) {
		m.Title = title
	}
}

func WithTarget(target string) Option {
	return func(m *Engine) {
		m.Target = target
	}
}

func WithFlags(flags Flags) Option {
	return func(m *Engine) {
		m.Flags = flags
	}
}

func WithChroma(chroma string) Option {
	return func(m *Engine) {
		m.Chroma = chroma
	}
}

func (f Flags) String() string {
	return fmt.Sprintf("%08b", f)
}

func modeName(flags Flags) string {
	return "ANSI"
}

func (ng *Engine) Run() error {
	if ng.Target == "" {
		ng.Target = "-"
	}

	slog.Debug("engine",
		"source", ng.Source,
		"target", ng.Target,
		"flags", ng.Flags.String(),
		"mode", modeName(ng.Flags),
		"chroma", ng.Chroma)

	b, err := readFile(ng.Source)
	if err != nil {
		return err
	}

	w, err := getWriter(ng.Target)
	if err != nil {
		return err
	}

	return ansi(w, b)
}

func ansi(w io.Writer, markdown []byte) error {
	out, err := glamour.RenderBytes(markdown, "dark")
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return nil
}

func readFile(src string) ([]byte, error) {
	if src == "-" {
		return io.ReadAll(os.Stdin)
	}
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func getWriter(target string) (io.Writer, error) {
	if target == "-" {
		return os.Stdout, nil
	}
	f, err := os.Create(target)
	if err != nil {
		return nil, fmt.Errorf("create target file: %w", err)
	}
	return f, nil
}
