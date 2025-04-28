package mdhero

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/charmbracelet/glamour"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
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
	Stdin    io.Reader
	Stdout   io.Writer
	Title    string
	Source   string
	Target   string
	Flags    Flags
	Chroma   string
	DiskPath string
}

func New(source string, options ...Option) *Engine {
	m := &Engine{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Source: source,
		Chroma: "tokyonight-moon",
	}

	for _, opt := range options {
		opt(m)
	}

	return m
}

type Flags uint8

func (f Flags) String() string {
	return fmt.Sprintf("%08b", f)
}

const (
	DEBUG   Flags = 1 << iota // debug mode
	HTML                      // output HTML
	BROWSER                   // open in browser
)

type Option func(*Engine)

func WithStdIO(r io.Reader, w io.Writer) Option {
	return func(m *Engine) {
		m.Stdin = r
		m.Stdout = w
	}
}

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

	if ng.Flags&DEBUG != 0 {
		fmt.Fprintf(os.Stderr, "%+v\n", ng)
	}

	b, err := ng.sourceReader()
	if err != nil {
		return err
	}

	w, err := ng.targetWriter()
	if err != nil {
		return err
	}

	var sink func(io.Writer, []byte) error

	if ng.Flags&HTML == HTML {
		sink = ng.html
	} else {
		sink = ng.ansi
	}

	if err := sink(w, b); err != nil {
		return err
	}

	if ok := HTML | BROWSER; ng.Flags&ok == ok {
		fmt.Fprintf(os.Stderr, "attempting to open %s in the browser...\n", ng.Target)
		openBrowser(ng.Target)
	}

	return nil
}

func (ng *Engine) ansi(w io.Writer, markdown []byte) error {
	out, err := glamour.RenderBytes(markdown, "dark")
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return nil
}

func (ng *Engine) html(w io.Writer, markdown []byte) error {
	var (
		title = defaultingTitle(ng.Title, ng.Source)
		gm    = newGoldmark(ng.Chroma)
		buf   bytes.Buffer
	)

	if err := gm.Convert(markdown, &buf); err != nil {
		return err
	}

	tpl := template.Must(template.New(title).Parse(pageTemplate))

	return tpl.Execute(w, pageContext{
		Title:   title,
		Content: buf.String(),
	})
}

func (ng *Engine) sourceReader() ([]byte, error) {
	if ng.Source == "-" {
		return io.ReadAll(ng.Stdin)
	}
	f, err := os.Open(ng.Source)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (ng *Engine) targetWriter() (io.Writer, error) {
	if ng.Target == "-" {
		return ng.Stdout, nil
	}
	f, err := os.Create(ng.Target)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "writing %s\n", ng.Target)
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

const pageTemplate = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
<style>
a.anchor { color: inherit; opacity: 0%; position: absolute; left: -28px; padding-right: 16px; }
:has( > a.anchor):hover a.anchor { opacity: 100%; }
:is(h1, h2, h3, h4, h5, h6):before { content: " "; width: 3rem; height: 1rem; }
body { background-color: #1a1b26; color: #c8d3f5; }
code {  background-color: #222436; color: #4fd6be; }
.content :is(h1, h2, h3, h4, h5, h6) { color: inherit; }
pre:has(code) { border-radius: 7px; }
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

func openBrowser(target string) error {
	return exec.Command("open", target).Run()
}
