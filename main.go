package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// global vars
var tmplPath string = "tmpl/"
var dataPath string = "data/"

//cache templates
var templates = template.Must(template.ParseFiles(tmplPath+"header.html", tmplPath+"nav.html", tmplPath+"footer.html", tmplPath+"home.html", tmplPath+"edit.html", tmplPath+"view.html", tmplPath+"viewAll.html"))

//router regex validation
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

var txtReplacer = strings.NewReplacer(".txt", "")

// page define
type Page struct {
	Title string
	Body  []byte
}

func addDataPath(p string) string {
	return dataPath + p
}

// page methods
func (p *Page) save() error {
	filename := addDataPath(p.Title + ".txt")
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// load page data
func loadPage(title string) (*Page, error) {
	filename := addDataPath(title + ".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/viewAll", viewAllHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/add/", addHandler)
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		fn(w, r, m[2])
	}
}

// handlers
func homeHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{Title: "Home"}
	renderTemplate(w, "home", p)
}

func viewAllHandler(w http.ResponseWriter, r *http.Request) {

	var allPages []string

	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		allPages = append(allPages, txtReplacer.Replace(f.Name()))
	}

	err = templates.ExecuteTemplate(w, "viewAll.html", allPages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func addHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		panic(err)
	}

	title := r.Form.Get("title")

	log.Printf(title)

	p := &Page{Title: title, Body: nil}
	err = p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/edit/"+p.Title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// render function
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
