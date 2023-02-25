// Package markdown provides markdown render functions.
package markdown

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/shurcooL/github_flavored_markdown"
)

// Render markdown from untrusted sources.
func Render(input string) string {
	// Render Markdown to HTML.
	html := github_flavored_markdown.Markdown([]byte(input))

	// Sanitize the HTML from any nasties.
	p := bluemonday.UGCPolicy()
	safened := p.SanitizeBytes(html)
	return string(safened)
}

// Quotify a message putting it into a Markdown "> quotes" block.
func Quotify(input string) string {
	var lines = []string{}
	for _, line := range strings.Split(input, "\n") {
		lines = append(lines, "> "+line)
	}
	return strings.Join(lines, "\n")
}
