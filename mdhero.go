package mdhero

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"text/template"

	"github.com/charmbracelet/glamour"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/anchor"
	"go.abhg.dev/goldmark/mermaid"
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
	HTML
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
	if flags&HTML != 0 {
		return "HTML"
	}
	return "ANSI"
}

func (ng *Engine) Run() error {
	if ng.Target == "" && ng.Flags&HTML == 0 {
		ng.Target = "-"
	}

	if ng.Target == "" {
		if ng.Source == "-" {
			ng.Target = "-"
		} else {
			ng.Target = strings.TrimSuffix(ng.Source, ".md") + ".html"
		}
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

	if ng.Flags&HTML == 0 {
		return ansi(w, b)
	}

	gm := newGoldmark(ng.Chroma)

	var buf bytes.Buffer

	if err := gm.Convert(b, &buf); err != nil {
		return err
	}

	return html(w, defaultingTitle(ng.Title, ng.Source), buf.String())
}

func ansi(w io.Writer, markdown []byte) error {
	out, err := glamour.RenderBytes(markdown, "dark")
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return nil
}

func html(w io.Writer, title, content string) error {
	tpl := template.Must(template.New("page").Parse(pageTemplate))
	return tpl.Execute(w, pageContext{
		Title:   title,
		Content: content,
	})
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

func newGoldmark(codeStyle string) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			extension.GFM,
			highlighting.NewHighlighting(highlighting.WithStyle(codeStyle)),
			&mermaid.Extender{},
			&anchor.Extender{Texter: anchor.Text("#"), Position: anchor.Before},
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)
}

func defaultingTitle(title, src string) string {
	if title != "" {
		return title
	}
	if src == "" {
		return "Markdown Hero"
	}
	name := strings.TrimSuffix(src, ".md")
	return name
}

type pageContext struct {
	Title   string
	Content string
}

const pageTemplate = `<!DOCTYPE html lANG="en">
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
<style>
:is(h1, h2, h3, h4, h5, h6) {
  position: relative;
  left: -28;
}
a.anchor { 
  color: inherit; 
  visibility: hidden;
}
:has( > a.anchor):hover a.anchor { 
  visibility: visible; 
}
</style>
</head>
<body>
<section class="section">
<div class="container is-max-tablet">
<div class="content">
{{.Content}}
</div>
</div>
</section>
</body>
</html>
`
