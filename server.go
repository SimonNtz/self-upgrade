package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type specifiHandler struct {
	addr     string
	ln       net.Listener
	server   *http.Server
	page     *template.Template
	signalCh chan os.Signal
}

const Version = "ver1"
const UpdateDir = "./dist/"

var (
	// Version        string // Defined by build flag with ```go build -ldflags="-X 'main.Version=vX'"````
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

func (sh *specifiHandler) handler(w http.ResponseWriter, r *http.Request) {
	if err := sh.page.Execute(w, Status); err != nil {
		log.Fatal(err)
	}
}

func (sh *specifiHandler) handlerCheck(w http.ResponseWriter, r *http.Request) {
	if newVersion := CheckNewVersion(); newVersion != "" {
		Status.NewVersion = newVersion
	}
	if err := sh.page.Execute(w, Status); err != nil {
		log.Fatal(err)
	}
}

func (sh *specifiHandler) handlerInstall(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", 302)
	fmt.Printf("reload signal received for exec %s.\n", NewVersionName)

	// DownloadFile(NewVersionName)
	// fmt.Printf("Exec downloaded.\n")

	p, err := UpdateExec(sh.addr, sh.ln, NewVersionName)
	if err != nil {
		fmt.Printf("Unable to start new executable: %v.\n", err)
	}
	fmt.Printf("New executable started pid:  %v.\n", p.Pid)
	// Create a context that will expire in 5 seconds and use this as a
	// timeout to Shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Return any errors during shutdown.
	err = sh.server.Shutdown(ctx)
	if err != nil {
		fmt.Fprintf(w, "Reload failed")
	}
	time.Sleep(100 * time.Millisecond)
}

func startServer(addr string, ln net.Listener) *http.Server {

	sh := specifiHandler{addr, ln, nil, nil}

	http.HandleFunc("/", sh.handler)
	http.HandleFunc("/check", sh.handlerCheck)
	http.HandleFunc("/install", sh.handlerInstall)

	httpServer := &http.Server{
		Addr: addr,
	}

	page, err := template.New("page").Parse(pageTemplate)
	if err != nil {
		log.Fatal(err)
	}

	sh.server = httpServer
	sh.page = page

	go httpServer.Serve(ln)

	<-sh.signalCh

	return httpServer
}

// Download file from local storage dir 'Dist'
// Should be passed an interface io.Reader for testing
// No error returned on Donwload
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
	// A FileMode represents a file's mode and permission bits. 770 - Owner and Group have all, and Other can read and execute
	to, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0770)
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

func main() {
	// Parse command line flags for the address to listen on.
	fmt.Printf(string(os.Getpid()) + "\n")
	var addr string
	flag.StringVar(&addr, "addr", ":9000", "Address to listen on.")
	// Create (or import) a net.Listener and start a goroutine that runs
	// a HTTP server on that net.Listener.
	ln, err := createOrImportListener(addr)
	if err != nil {
		fmt.Printf("Unable to create or import a listener: %v.\n", err)
		os.Exit(1)
	}
	startServer(addr, ln)
	// Wait for signals to either fork or quit.
	fmt.Printf("Exiting.\n")
}
