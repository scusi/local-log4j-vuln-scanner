/* FileUpload.go - starts a webserver with an upload form to upload files.

Usage:

 fileUpload -addr=:8123 -uploadDir=/tmp/upload -showAssets=true

The above command would start a webserver on port 8080, allow uploads to
/tmp/upload and shows the content of this directory.

 fileUpload -addr=:8443 -sslOn=true -cert=cert.pem -key=key.pem -uploadDir=/tmp/upload -showAssets=true

The above command would start a TLS/SSL Webserver on port 8443, using key.pem
as key and cert.pem as cert. Uploaded files fill go to /tmp/upload.
Uploaded files will be shown.

cert.pem must be a pem encoded x509 certificate for webservers, and key.pem
must be the corresponding key, also pem encoded.

*/

package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// supported command line flags
var addr = flag.String("addr", "127.0.0.1:8123", "http service address")
var uploadDir = flag.String("uploadDir", "./", "directory to upload files to")
var templateDir = flag.String("templateDir", "", "directory with templates, relative to working directory")
var assetDir = flag.Bool("showAssets", true, "if set to 'true' uploadDir will be browsable unter '/files/'")
var sslOn = flag.Bool("sslOn", false, "if set to 'true' HTTPS is turned on.")
var cert = flag.String("cert", "cert.pem", "pem encoded certificate to use")
var key = flag.String("key", "key.pem", "pem encoded key to use")

// build in templates (default)
const indexDefaultTemplate = `
<!DOCTYPE html>
<html>
 <head>
  <title>FileServer</title>
  <link type="text/css" rel="stylesheet" href="/css/style.css" />
  <script type="text/javascript">
    function resizeIframe(obj){
       obj.style.height = 0;
       obj.style.height = obj.contentWindow.document.body.scrollHeight + 'px';
    }
 </script>
 </head>
 <body>
  <h1>FileServer</h1>
  <ul>
   <li><a href="/">Home</a></li>
   <li><a href="/files">Files</a></li>
   <li><a href="/upload">Upload</a></li>
  </ul>
  <h2>Files</h2>
  <iframe onload='resizeIframe(this)' src="/files/" width="90%" heigth="200">
  </iframe>
 </body>
</html>
`

const uploadDefaultTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>File Upload Demo</title>
    <link type="text/css" rel="stylesheet" href="/css/style.css" />
  </head>
  <body>
    <div class="container">
  <ul>
   <li><a href="/">Home</a></li>
   <li><a href="/files">Files</a></li>
   <li><a href="/upload">Upload</a></li>
  </ul>
    </div>
    <div class="container">
      <h1>File Upload</h1>
      <div class="message">{{.}}</div>
      <form class="form-signin" method="post" action="/upload" enctype="multipart/form-data">
          <fieldset>
            <input type="file" name="myfiles" id="myfiles" multiple="multiple">
            <input type="submit" name="submit" value="Submit">
        </fieldset>
      </form>
    </div>
  </body>
</html>

`

var templates *template.Template

func init() {
	// Parse HTML templates
	// template files are excpected to be in "tmpl/" relative to the working directory.
	// Working directory is the directory where you start FileUpload from.
	// derectory can be overwritten by useing the flag 'templateDir' on the command line.
	if *templateDir != "" {
		templates = template.Must(template.ParseFiles(*templateDir+"/upload.html", *templateDir+"/index.html"))
	} else {
		var tmpl *template.Template
		tmpl, err := template.New("index.html").Parse(indexDefaultTemplate)
		if err != nil {
			log.Fatal(err)
		}
		tmpl, err = tmpl.New("upload.html").Parse(uploadDefaultTemplate)
		if err != nil {
			log.Fatal(err)
		}
		templates = tmpl
	}
}

//Display the named template
func display(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		log.Printf("display: error execute template: %s\n", err.Error())
	}
}

// Show index
func indexHandler(w http.ResponseWriter, r *http.Request) {
	//log.Printf("indexHandler started\n")
	display(w, "index", nil)
}

//This is where the action happens.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	//GET displays the upload form.
	case "GET":
		display(w, "upload", nil)
		log.Printf("show upload form to %s\n", r.RemoteAddr)

	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		//get the multipart reader for the request.
		reader, err := r.MultipartReader()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//copy each part to destination.
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			//if part.FileName() is empty, skip this iteration.
			if part.FileName() == "" {
				continue
			}
			// TODO: Add Input Validation
			// Make sure uploadDir is existing or create it
			*uploadDir = filepath.Clean(*uploadDir)
			err = os.MkdirAll(*uploadDir, 0755)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			dst, err := os.Create(filepath.Join(*uploadDir, part.FileName()))
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("file %s created.\n", filepath.Join(*uploadDir, part.FileName()))
			if _, err := io.Copy(dst, part); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("uploaded content copied to %s.\n", filepath.Join(*uploadDir, part.FileName()))
			//log.Printf("upload of %s (%d bytes) by %s \n", part.FileName(), b, r.RemoteAddr)
			log.Printf("upload of %s by %s \n", part.FileName(), r.RemoteAddr)
		}
		//display success message.
		display(w, "upload", "Upload successful.")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)

	// if 'assetDir' is set to 'true' uploadDir will be browsable via '/assets/'
	if *assetDir == true {
		//static file handler.
		http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(*uploadDir))))
		//http.Handle("/", http.FileServer(http.Dir(*uploadDir)))
	}

	if *sslOn != true {
		log.Printf("Listening on %s\n", *addr)
		log.Printf("point your browser to: https://localhost%s/\n", *addr)
		err := http.ListenAndServe(*addr, nil)
		if err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	} else {
		log.Printf("Setup TLS with %s as key and %s as cert\n", *key, *cert)
		log.Printf("Listening on %s\n", *addr)
		log.Printf("point your browser to: https://localhost%s/\n", *addr)
		err := http.ListenAndServeTLS(*addr, *cert, *key, nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS:", err)
		}
	}
}

func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s\t%s",
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

