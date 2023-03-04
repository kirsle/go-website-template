package templates

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aichaos/silhouette/webapp/config"
	"github.com/aichaos/silhouette/webapp/markdown"
	"github.com/aichaos/silhouette/webapp/session"
	"github.com/aichaos/silhouette/webapp/utility"
)

// Generics
type Number interface {
	int | int64 | uint64 | float32 | float64
}

// TemplateFuncs available to all pages.
func TemplateFuncs(r *http.Request) template.FuncMap {
	return template.FuncMap{
		"InputCSRF":         InputCSRF(r),
		"SincePrettyCoarse": SincePrettyCoarse(),
		"ComputeAge":        utility.Age,
		"Split":             strings.Split,
		"ToMarkdown":        ToMarkdown,
		"ToJSON":            ToJSON,
		"Now":               time.Now,
		"Pluralize":         Pluralize[int],
		"Pluralize64":       Pluralize[int64],
		"PluralizeU64":      Pluralize[uint64],
		"Substring":         Substring,
		"TrimEllipses":      TrimEllipses,
		"IterRange":         IterRange,
		"SubtractInt":       SubtractInt,
		"UrlEncode":         UrlEncode,
		"QueryPlus":         QueryPlus(r),
	}
}

// InputCSRF returns the HTML snippet for a CSRF token hidden input field.
func InputCSRF(r *http.Request) func() template.HTML {
	return func() template.HTML {
		ctx := r.Context()
		if token, ok := ctx.Value(session.CSRFKey).(string); ok {
			return template.HTML(fmt.Sprintf(
				`<input type="hidden" name="%s" value="%s">`,
				config.CSRFInputName,
				token,
			))
		} else {
			return template.HTML(`[CSRF middleware error]`)
		}
	}
}

// SincePrettyCoarse formats a time.Duration in plain English. Intended for "joined 2 months ago" type
// strings - returns the coarsest level of granularity.
func SincePrettyCoarse() func(time.Time) template.HTML {
	return func(since time.Time) template.HTML {
		return template.HTML(utility.FormatDurationCoarse(time.Since(since)))
	}
}

// ToMarkdown renders input text as Markdown.
func ToMarkdown(input string) template.HTML {
	return template.HTML(markdown.Render(input))
}

// ToJSON will stringify any json-serializable object.
func ToJSON(v any) template.JS {
	bin, err := json.Marshal(v)
	if err != nil {
		return template.JS(err.Error())
	}
	return template.JS(string(bin))
}

// Pluralize text based on a quantity number. Provide up to 2 labels for the
// singular and plural cases, or the defaults are "", "s"
func Pluralize[V Number](count V, labels ...string) string {
	if len(labels) < 2 {
		labels = []string{"", "s"}
	}

	if count == 1 {
		return labels[0]
	}
	return labels[1]
}

// Substring safely returns the first N characters of a string.
func Substring(value string, n int) string {
	if n > len(value) {
		return value
	}
	return value[:n]
}

// TrimEllipses is like Substring but will add an ellipses if truncated.
func TrimEllipses(value string, n int) string {
	if n > len(value) {
		return value
	}
	return value[:n] + "…"
}

// IterRange returns a list of integers useful for pagination.
func IterRange(start, n int) []int {
	var result = []int{}
	for i := start; i <= n; i++ {
		result = append(result, i)
	}
	return result
}

// SubtractInt subtracts two numbers.
func SubtractInt(a, b int) int {
	return a - b
}

// UrlEncode escapes a series of values (joined with no delimiter)
func UrlEncode(values ...interface{}) string {
	var result string
	for _, value := range values {
		result += url.QueryEscape(fmt.Sprintf("%v", value))
	}
	return result
}

// QueryPlus takes the current request's query parameters and upserts them with new values.
//
// Use it like: {{QueryPlus "page" .NextPage}}
//
// Returns the query string sans the ? prefix, like "key1=value1&key2=value2"
func QueryPlus(r *http.Request) func(...interface{}) template.URL {
	return func(upsert ...interface{}) template.URL {
		// Get current parameters.
		var params = r.Form

		// Mix in the incoming fields.
		for i := 0; i < len(upsert); i += 2 {
			var (
				key   = fmt.Sprintf("%v", upsert[i])
				value interface{}
			)
			if len(upsert) > i {
				value = upsert[i+1]
			}

			params[key] = []string{fmt.Sprintf("%v", value)}
		}

		// Assemble and return the query string.
		var parts = []string{}
		for k, vs := range params {
			for _, v := range vs {
				parts = append(parts,
					fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)),
				)
			}
		}
		return template.URL(strings.Join(parts, "&"))
	}
}
