package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
)

type templates map[string]*template.Template

type author struct {
	ID   string
	Name string
}

func (a author) uri() string {
	return fmt.Sprintf("/authors/%s", a.ID)
}

func (a author) is(other author) bool {
	return a.ID == other.ID
}

type snippet struct {
	Author   author
	PostedAt time.Time
	Body     string
}

func (s snippet) toRenderableSnippet() renderableSnippet {
	return renderableSnippet{
		PostedAt:   humanize.Time(s.PostedAt),
		Body:       s.Body,
		AuthorName: s.Author.Name,
		AuthorURI:  s.Author.uri(),
	}
}

type renderableSnippet struct {
	PostedAt   string
	Body       string
	AuthorName string
	AuthorURI  string
}

type snippetsPage struct {
	Snippets []renderableSnippet
}

type snippetsRepo interface {
	list() []snippet
}

type inMemorySnippetsRepo struct {
	snippets []snippet
}

func (r inMemorySnippetsRepo) list() []snippet {
	return r.snippets
}

func (r inMemorySnippetsRepo) listByAuthor(a author) []snippet {
	snippetsByAuthor := []snippet{}
	for _, s := range r.snippets {
		if a.is(s.Author) {
			snippetsByAuthor = append(snippetsByAuthor, s)
		}
	}
	return snippetsByAuthor
}

type authorRepo interface {
	list() []author
	getById(id string) *author
}

type inMemoryAuthorRepo struct {
	authors []author
}

func (r inMemoryAuthorRepo) list() []author {
	return r.authors
}

func (r inMemoryAuthorRepo) getById(id string) *author {
	for _, a := range r.authors {
		if a.ID == id {
			return &a
		}
	}
	return nil
}

func makeViewSnippetsHandler(tmpl templates, sr snippetsRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snippets := sr.list()
		snippetsPage := snippetsPage{}
		for _, s := range snippets {
			snippetsPage.Snippets = append(snippetsPage.Snippets, s.toRenderableSnippet())
		}

		tmpl["snippets.html"].ExecuteTemplate(w, "base", snippetsPage)
	}
}

func makeViewSnippetsByAuthorHandler(tmpl templates, ar authorRepo, sr snippetsRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/authors/"):]
		a := ar.getById(id)
		if a == nil {
			panic("nil author")
		}

		snippets := sr.list()
		snippetsPage := snippetsPage{}
		for _, s := range snippets {
			snippetsPage.Snippets = append(snippetsPage.Snippets, s.toRenderableSnippet())
		}

		tmpl["snippets.html"].ExecuteTemplate(w, "base", snippetsPage)
	}
}

func main() {
	tmpl := make(map[string]*template.Template)
	tmpl["snippets.html"] = template.Must(template.ParseFiles("views/snippets.html", "views/base.html"))

	a := author{Name: "christian scott", ID: "0"}

	snippets := []snippet{
		{
			Author:   a,
			PostedAt: time.Now(),
			Body:     "I worked on this shitty snippets tool",
		},
	}
	snippetsRepo := inMemorySnippetsRepo{snippets}

	authors := []author{a}
	authorsRepo := inMemoryAuthorRepo{authors}

	http.HandleFunc("/", makeViewSnippetsHandler(tmpl, snippetsRepo))
	http.HandleFunc("/authors/", makeViewSnippetsByAuthorHandler(tmpl, authorsRepo, snippetsRepo))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
