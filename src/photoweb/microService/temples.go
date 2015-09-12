package microService

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
)

const DEFAULT_TEMPLATE_DIR = "./views"

type Templates struct {
	templates map[string]*template.Template
}

func NewTemplates(templateDir string) *Templates {
	var templates = &Templates{templates: make(map[string]*template.Template)}
	templates.LoadTemplates(templateDir)
	return templates
}

func (this *Templates) LoadTemplates(templateDir string) {
	if templateDir == "" {
		templateDir = DEFAULT_TEMPLATE_DIR
	}
	fileInfoArr, err := ioutil.ReadDir(templateDir)
	if err != nil {
		log.Fatal("LoadTemplates: " + err.Error())
		return
	}
	for _, fileInfo := range fileInfoArr {
		templateName := fileInfo.Name()
		if path.Ext(templateName) != ".html" {
			continue
		}
		templatePath := TEMPLATE_DIR + "/" + templateName
		log.Println("Loading template: " + templatePath)
		t := template.Must(template.ParseFiles(templatePath))
		this.templates[templateName] = t
	}
}

func (this *Templates) renderTemplate(rw http.ResponseWriter, tmpl string, locals map[string]interface{}) (err error) {
	/*
	    var t *template.Template
	    t, err = template.ParseFiles("views/" + tmpl + ".html")
	    if err != nil {
	        return
	    }
	    err = t.Execute(rw, locals)
	    return
	//*/
	err = this.templates[tmpl+".html"].Execute(rw, locals)
	return
}
