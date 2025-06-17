# Markdown Hero

A command line tool to work with markdown files.

## Why

While markdown is great, and its great that most web sites support it,
it is not so easy to render and view markdown files locally, when they
contain rich content like diagrams, code blocks, and other things.

There are a number of tools that can be combined to do this, but they
are not always easy to use, and they are not always available on all
platforms.

This tool is designed provide the full set of required features to
effectively work with markdown files.

## Installation

```bash
go install github.com/bluebrown/markdown-hero/cmd/md@latest
```

## Usage

```bash
md [-html] [-browser] <source> [<target>]
```

### Mode

By default, the tool will render the markdown file in `ANSI mode`. `HTML
mode` is enabled by the `-html` flag.

Depending on the mode an empty target has different semantics.

- In ANSI mode, the target is set to `-`, which means write to stdout.
- In HTML mode, the target is set to the source file with a `.html`
  extension.

### Browser

If the `-browser` flag is set, the generated HTML file is opened in the
browser. If `-html` is not explicitly set alongside, HTML output is
assumed, but the generated file path defaults to os temporary directory,
instead of following the source file name.

The browser is opened on a best effort basis, and may not work on all
platforms. If the browser fails to open, the HTML file will still be
created, and can be opened manually.
