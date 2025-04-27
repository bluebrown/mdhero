# Markdown Hero

A command line tool to work with markdown files.

## Why

While markdown is great, and its great that most web sites support it, it is
not so easy to render and view markdown files locally, when they contain rich
content like diagrams, code blocks, and other things.

There are a number of tools that can be combined to do this, but they are not
always easy to use, and they are not always available on all platforms.

This tool is designed provide the full set of required features to effectively
work with markdown files.

## Features

- [x] ANSI Rendering
- [x] HTML Rendering
  - [x] Mermaid diagrams
  - [x] Heading Anchors
  - [ ] TOC Generation
- [x] Github flavored markdown
- [x] Syntax highlighting
- [ ] Limited Styling
- [ ] Live preview
- [ ] Directory Mode
- [ ] Remote Sources

## Installation

```bash
go install github.com/bluebrown/markdown-hero/cmd/md@latest
```

## Usage

```bash
md [-debug] [-html] <source>
```
