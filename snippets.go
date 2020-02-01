package main

import (
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
)

type templates map[string]*template.Template

type snippet struct {
	PostedAt time.Time
	Body     string
}

type renderableSnippet struct {
	PostedAt string
	Body     string
}

type snippetsPage struct {
	Snippets []renderableSnippet
}

func makeViewSnippetsHandler(tmpl templates, getSnippets func() []snippet) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snippets := getSnippets()
		snippetsPage := snippetsPage{}
		for _, s := range snippets {
			snippetsPage.Snippets = append(snippetsPage.Snippets, renderableSnippet{
				PostedAt: humanize.Time(s.PostedAt),
				Body:     s.Body,
			})
		}

		tmpl["snippets.html"].ExecuteTemplate(w, "base", snippetsPage)
	}
}

func main() {
	tmpl := make(map[string]*template.Template)
	tmpl["snippets.html"] = template.Must(template.ParseFiles("views/snippets.html", "views/base.html"))

	snippets := []snippet{
		snippet{
			PostedAt: time.Now(),
			Body:     "I worked on this shitty snippets tool",
		},
	}
	getSnippets := func() []snippet {
		return snippets
	}

	http.HandleFunc("/", makeViewSnippetsHandler(tmpl, getSnippets))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
