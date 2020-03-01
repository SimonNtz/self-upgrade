package main

import (
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const Version = "ver1"

var (
	pageTemplate = `
<!DOCTYPE html>
<html>
<head>
<title> Server {{.Version}} </title>
</head>
<body>
<h1>This server is version {{.Version}}</h1>
<a href="check">Check for new version</a>
<br>
{{if .NewVersion}}New version is available: {{.NewVersion}} | <a
href="install">Upgrade</a>{{end}}
</body>
</html>
`
	Status = struct{ Version, NewVersion string }{Version, ""}
)

func main() {
	page, err := template.New("page").Parse(pageTemplate)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := page.Execute(w, Status); err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		if newVersion := CheckNewVersion(); newVersion != "" {
			Status.NewVersion = newVersion
		}
		if err := page.Execute(w, Status); err != nil {
			log.Fatal(err)
		}
	})
	http.HandleFunc("/install", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Not implemented %v", html.EscapeString(r.URL.Path))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// HARDCOEDED PATH

func CheckNewVersion() string {
	if filesName := listDir(); filesName != nil {
		return strings.Split(filesName[0], ".")[1]
	}
	return ""
}

func listDir() (filesName []string) {
	files, err := ioutil.ReadDir("./dist")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		filesName = append(filesName, f.Name())
	}
	return
}
