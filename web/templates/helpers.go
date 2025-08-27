package templates

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var markdownParser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
		html.WithUnsafe(), // Allow raw HTML in markdown
	),
)

// RenderMarkdown converts markdown text to HTML
func RenderMarkdown(markdown string) template.HTML {
	if markdown == "" {
		return template.HTML("")
	}

	var buf bytes.Buffer
	if err := markdownParser.Convert([]byte(markdown), &buf); err != nil {
		// Return the raw markdown if parsing fails
		return template.HTML("<pre>" + template.HTMLEscapeString(markdown) + "</pre>")
	}

	return template.HTML(buf.String())
}