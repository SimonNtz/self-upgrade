package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// const Version = "ver1"
const UpdateDir = "./dist/"

var (
	Version        string // Defined by build flag with ```go build -ldflags="-X 'main.Version=vX'"````
	NewVersionName string
	pageTemplate   = `
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
		DownloadFile(NewVersionName)
		fmt.Fprintf(w, "Installing update: %s", NewVersionName)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Download file from local storage dir 'Dist'
// Should be passed an interface io.Reader for testing
func DownloadFile(filename string) {

	// Open the file that should be copied
	// Read the contents
	// Create and open the file that the contents should be copied into
	// Write to the new file
	// Close both files

	from, err := os.Open(UpdateDir + filename)
	if err != nil {
		log.Fatal(err)
	}
	defer from.Close()

	to, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatal(err)
	}
}

// TODO: pass as an argument HARDCODED DIR '/DIST' PATH
// Check dir exists
// Assumption take first update sorted by date

func CheckNewVersion() string {
	if filesName := listDir(); filesName != nil {
		NewVersionName = filesName[0]
		return strings.Split(filesName[0], ".")[1]
	}
	return ""
}

func listDir() (filesName []string) {
	files, err := ioutil.ReadDir(UpdateDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		filesName = append(filesName, f.Name())
	}
	return
}
