package models

import (
	"math"
	"net/http"
	"strconv"

	"github.com/aichaos/silhouette/webapp/config"
)

// Pagination result object.
type Pagination struct {
	Page    int // provide <0 to mean "last page"
	PerPage int
	Total   int64
	Sort    string

	// privates
	lastPage bool
}

// Page for Iter.
type Page struct {
	Page      int
	IsCurrent bool
}

// Load the page from form or query parameters.
func (p *Pagination) ParsePage(r *http.Request) {
	raw := r.FormValue("page")
	a, err := strconv.Atoi(raw)
	if err == nil {
		if a <= 0 {
			p.lastPage = true
			a = 1
		}
		p.Page = a
	} else {
		p.Page = 1
	}
}

// Iter the pages, for templates.
func (p *Pagination) Iter() []Page {
	var (
		pages   = []Page{}
		pageIdx int
		total   = p.Pages()
	)
	for i := 1; i <= total; i++ {
		pages = append(pages, Page{
			Page:      i,
			IsCurrent: i == p.Page,
		})

		if i == p.Page {
			pageIdx = i
		}
	}

	// Do we have A LOT of pages?
	if len(pages) > config.PagerButtonLimit {
		// We return a slide only N pages long. Where is our current page in the offset?
		if pageIdx <= config.PagerButtonLimit/2 {
			// We are near the front, return the first N pages.
			return pages[:config.PagerButtonLimit+1]
		}

		// Are we near the end?
		if pageIdx > len(pages)-(config.PagerButtonLimit/2) {
			// We are near the end, return the last N pages.
			return pages[len(pages)-config.PagerButtonLimit-1:]
		}

		// We are somewhere in the middle.
		var result = []Page{}
		for i := pageIdx - (config.PagerButtonLimit / 2) - 1; i < pageIdx+(config.PagerButtonLimit/2); i++ {
			if i >= 0 && i < len(pages) {
				result = append(result, pages[i])
			}
		}
		return result
	}
	return pages
}

func (p *Pagination) Pages() int {
	if p.PerPage == 0 {
		return 0
	}
	return int(math.Ceil(float64(p.Total) / float64(p.PerPage)))
}

func (p *Pagination) GetOffset() int {
	// Are we looking for the FINAL page?
	if p.lastPage && p.Pages() >= 1 {
		p.Page = p.Pages()
	}
	return (p.Page - 1) * p.PerPage
}

func (p *Pagination) HasNext() bool {
	return p.Page < p.Pages()
}

func (p *Pagination) HasPrevious() bool {
	return p.Page > 1
}

func (p *Pagination) Next() int {
	if p.Page >= p.Pages() {
		return p.Pages()
	}
	return p.Page + 1
}

func (p *Pagination) Previous() int {
	if p.Page > 1 {
		return p.Page - 1
	}
	return 1
}
