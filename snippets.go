package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
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
	ShowNewSnippetForm bool
	Snippets           []renderableSnippet
}

type snippetsRepo interface {
	list() []snippet
	listByAuthor(a author) []snippet
	add(s snippet) error
}

type inMemorySnippetsRepo struct {
	sync.RWMutex
	snippets []snippet
}

func newInMemorySnippetsRepo(snippets []snippet) inMemorySnippetsRepo {
	return inMemorySnippetsRepo{sync.RWMutex{}, snippets}
}

func (r *inMemorySnippetsRepo) list() []snippet {
	return r.snippets
}

func (r *inMemorySnippetsRepo) listByAuthor(a author) []snippet {
	r.Lock()
	defer r.Unlock()

	var snippetsByAuthor []snippet
	for _, s := range r.snippets {
		if a.is(s.Author) {
			snippetsByAuthor = append(snippetsByAuthor, s)
		}
	}
	return snippetsByAuthor
}

func (r *inMemorySnippetsRepo) add(s snippet) error {
	r.Lock()
	defer r.Unlock()

	r.snippets = append([]snippet{s}, r.snippets...)
	return nil
}

type authorRepo interface {
	list() []author
	getById(id string) *author
}

type inMemoryAuthorRepo struct {
	authors []author
}

func (r *inMemoryAuthorRepo) list() []author {
	return r.authors
}

func (r *inMemoryAuthorRepo) getById(id string) *author {
	for _, a := range r.authors {
		if a.ID == id {
			return &a
		}
	}
	return nil
}

func makeViewSnippetsHandler(tmpl templates, sr snippetsRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			must(r.ParseForm())
			s := snippet{
				Author: author{ID: "1", Name: "Someone New"},
				Body:   r.Form.Get("snippet"),
			}
			must(sr.add(s))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		snippets := sr.list()
		must(renderSnippets(tmpl, w, snippets, true))
	}
}

func makeViewSnippetsByAuthorHandler(tmpl templates, ar authorRepo, sr snippetsRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/authors/"):]
		a := ar.getById(id)
		if a == nil {
			panic("nil author")
		}

		snippets := sr.listByAuthor(*a)
		must(renderSnippets(tmpl, w, snippets, false))
	}
}

func renderSnippets(tmpl templates, w http.ResponseWriter, ss []snippet, showNewSnippetsForm bool) error {
	page := snippetsPage{ShowNewSnippetForm: showNewSnippetsForm}
	for _, s := range ss {
		page.Snippets = append(page.Snippets, s.toRenderableSnippet())
	}
	return tmpl["snippets.html"].ExecuteTemplate(w, "base", page)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	tmpl := make(map[string]*template.Template)
	tmpl["snippets.html"] = template.Must(template.ParseFiles("views/snippets.html", "views/base.html"))

	a := author{Name: "christian scott", ID: "0"}

	snippetsRepo := newInMemorySnippetsRepo([]snippet{})

	must(snippetsRepo.add(snippet{
		Author:   a,
		PostedAt: time.Now(),
		Body:     "I worked on this shitty snippets tool",
	}))

	authors := []author{a}
	authorsRepo := inMemoryAuthorRepo{authors}

	http.HandleFunc("/", makeViewSnippetsHandler(tmpl, &snippetsRepo))
	http.HandleFunc("/authors/", makeViewSnippetsByAuthorHandler(tmpl, &authorsRepo, &snippetsRepo))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
