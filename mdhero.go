package mdhero

//go:generate npm -C assets install
//go:generate npm -C assets run build

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

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
	Stdin  io.Reader
	Stdout io.Writer
	Title  string
	Source string
	Target string
	Flags  Flags
	Chroma string
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
	if ng.Flags&DEBUG != 0 {
		fmt.Fprintf(os.Stderr, "%+v\n", ng)
	}

	raw, err := ng.readSource()
	if err != nil {
		return err
	}

	var render func([]byte) ([]byte, error)

	if ng.Flags&(HTML|BROWSER) > 0 {
		render = ng.html
	} else {
		render = ng.ansi
	}

	res, err := render(raw)
	if err != nil {
		return err
	}

	if err := ng.writeTarget(res); err != nil {
		return err
	}

	if ng.Flags&BROWSER > 0 {
		fmt.Fprintf(os.Stderr, "attempting to open %s in the browser...\n", ng.Target)
		openBrowser(ng.Target, ng.Flags&DEBUG != 0)
	}

	return nil
}

func (ng *Engine) ansi(markdown []byte) ([]byte, error) {
	return glamour.RenderBytes(markdown, "dark")
}

func (ng *Engine) html(markdown []byte) ([]byte, error) {
	var (
		title = defaultingTitle(ng.Title, ng.Source)
		gm    = newGoldmark(ng.Chroma)
	)

	var md bytes.Buffer
	if err := gm.Convert(markdown, &md); err != nil {
		return nil, err
	}

	tpl := template.Must(template.New(title).Parse(pageTemplate))

	var htm bytes.Buffer
	err := tpl.Execute(&htm, pageContext{
		Title:   title,
		Style:   string(bulmaCSS),
		Content: md.String(),
	})

	return htm.Bytes(), err
}

func (ng *Engine) readSource() ([]byte, error) {
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

func (ng *Engine) configureTargetPath() {
	if ng.Target != "" {
		return
	}
	if ng.Flags&(HTML|BROWSER) == 0 {
		ng.Target = "-"
		return
	}
	if ng.Flags&HTML == 0 {
		ng.Target = filepath.Join(os.TempDir(), "mdhero.preview.html")
		return
	}
	ng.Target = strings.TrimSuffix(ng.Source, ".md") + ".html"

}

func (ng *Engine) writeTarget(data []byte) error {
	ng.configureTargetPath()

	var w io.Writer = ng.Stdout
	if ng.Target != "-" {
		f, err := os.Create(ng.Target)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	if ng.Target != "-" || ng.Flags&DEBUG != 0 {
		fmt.Fprintf(os.Stderr, "writing %s\n", ng.Target)
	}
	_, err := w.Write(data)
	return err
}

func newGoldmark(codeStyle string) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			extension.GFM,
			highlighting.NewHighlighting(highlighting.WithStyle(codeStyle)),
			&mermaid.Extender{Theme: "dark"},
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
	Style   string
	Content string
}

//go:embed assets/bulma.css
var bulmaCSS []byte

const pageTemplate = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
{{.Style}}
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

var browers = []string{
	os.Getenv("BROWSER"),
	os.Getenv("XDG_OPEN"),
	"xdg-open",
	"open",
	"start",
}

func openBrowser(target string, debug bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exe := anyExecutable(browers)
	if exe == "" {
		if debug {
			fmt.Fprintf(os.Stderr, "no browser found\n")
		}
		return nil
	}

	if err := exec.CommandContext(ctx, exe, target).Start(); err != nil && debug {
		fmt.Fprintf(os.Stderr, "failed to open browser: %v\n", err)
	}

	return nil
}

func anyExecutable(tryNames []string) string {
	for _, name := range tryNames {
		if name == "" {
			continue
		}
		if _, err := exec.LookPath(name); err != nil {
			continue
		}
		return name
	}
	return ""
}
