package microService

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime/debug"
)

const (
	UPLOAD_DIR   = "./uploads"
	TEMPLATE_DIR = "./views"
	ListDir      = 0x0001
)

type PhotoService struct {
	templates *Templates
}

func NewPhotoService() *PhotoService {
	return &PhotoService{NewTemplates(TEMPLATE_DIR)}
}

func (this *PhotoService) isExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (this *PhotoService) renderPage(rw http.ResponseWriter, tmpl string, locals map[string]interface{}) (err error) {
	return this.templates.renderTemplate(rw, tmpl, locals)
}

func (this *PhotoService) check(err error) {
	if err != nil {
		panic(err)
	}
}

func (this *PhotoService) safeHandler(fn func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(rw, e.Error(), http.StatusInternalServerError)
				log.Println("Error: panic in %v. - %v", fn, e)
				log.Println(string(debug.Stack()))
			}
		}()
		fn(rw, req)
	}
}

func (this *PhotoService) indexHandler(w http.ResponseWriter, r *http.Request) {
	this.renderPage(w, "index", nil)
}

func (this *PhotoService) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		this.renderPage(w, "upload", nil)
	} else if r.Method == "POST" {
		f, h, err := r.FormFile("image")
		this.check(err)
		filename := h.Filename
		defer f.Close()
		t, err := os.Create(UPLOAD_DIR + "/" + filename)
		this.check(err)
		defer t.Close()
		_, err = io.Copy(t, f)
		this.check(err)
		http.Redirect(w, r, "/view?id="+filename, http.StatusFound)
	}
}

func (this *PhotoService) viewHandler(rw http.ResponseWriter, req *http.Request) {
	imageId := req.FormValue("id")
	imagePath := UPLOAD_DIR + "/" + imageId
	if !this.isExists(imagePath) {
		http.NotFound(rw, req)
		return
	}
	rw.Header().Set("Content-Type", "image")
	http.ServeFile(rw, req, imagePath)
}

func (this *PhotoService) listHandler(w http.ResponseWriter, r *http.Request) {
	fileInfoArr, err := ioutil.ReadDir("./uploads")
	this.check(err)

	locals := make(map[string]interface{})
	images := []string{}
	for _, fileInfo := range fileInfoArr {
		images = append(images, fileInfo.Name())
	}
	locals["images"] = images
	/*
		t, err := template.ParseFiles("views/list.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, locals)
	*/
	this.renderPage(w, "list", locals)
}

func (this *PhotoService) staticDirHandler(mux *http.ServeMux, prefix string, staticDir string, flag int) {
	mux.HandleFunc(prefix, func(rw http.ResponseWriter, req *http.Request) {
		file := staticDir + req.URL.Path[len(prefix)-1:]
		log.Println("staticDirHandler : " + file)
		if (flag & ListDir) == 0 {
			if !this.isExists(file) {
				http.NotFound(rw, req)
				return
			}
		}
		http.ServeFile(rw, req, file)
	})
}

func (this *PhotoService) statusHandler(rw http.ResponseWriter, req *http.Request) {
	io.WriteString(rw, "running")
}

func (this *PhotoService) Start() {
	log.Println("Start photoweb server.") // sudo kill -9 `sudo ps aux|grep photoweb|awk {'print $2'}|sed -n '1,2p'`
	mux := http.NewServeMux()
	this.staticDirHandler(mux, "/assets/", "./public", 0)
	mux.HandleFunc("/", this.safeHandler(this.indexHandler))
	mux.HandleFunc("/upload", this.safeHandler(this.uploadHandler))
	mux.HandleFunc("/view", this.safeHandler(this.viewHandler))
	mux.HandleFunc("/list", this.safeHandler(this.listHandler))
	mux.HandleFunc("/status", this.safeHandler(this.statusHandler))

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
