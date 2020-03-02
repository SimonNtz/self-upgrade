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
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type specifiHandler struct {
	page     *template.Template
	signalCh chan os.Signal
}

const (
	Version            = "ver1"
	UpdateDir          = "dist"
	SignatureExtension = ".RSAsignature"
)

var (
	// Version        string // Defined by build flag with ```go build -ldflags="-X 'main.Version=vX'"````
	newVersionName string
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
	if newVersion := checkNewVersion(); newVersion != "" {
		Status.NewVersion = newVersion
	}
	if err := sh.page.Execute(w, Status); err != nil {
		log.Fatal(err)
	}
}

func (sh *specifiHandler) handlerInstall(w http.ResponseWriter, r *http.Request) {
	if newVersionName == "" {
		http.Error(w, `Unauthorized access`, http.StatusUnauthorized)
		return
	}

	err := DownloadAndVerifyFile(filepath.Join(UpdateDir, newVersionName))
	if err != nil {
		errString := fmt.Sprintf("Error while donwloading %s: %v", newVersionName, err)
		log.Fatalf(errString)
		http.Error(w, errString, http.StatusInternalServerError)
		return
	}

	defer http.Redirect(w, r, "/", 302)
	sh.signalCh <- syscall.SIGINT

}

func startServer(addr string, ln net.Listener) *http.Server {

	sh := specifiHandler{nil, nil}

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

	signalCh := make(chan os.Signal, 1024)

	go httpServer.Serve(ln)

	sh.page = page
	sh.signalCh = signalCh
	signal.Notify(signalCh, syscall.SIGHUP)

	<-signalCh

	p, err := RestartExec(addr, ln)
	if err != nil {
		fmt.Printf("Error while installing update: %s: %v.\n", newVersionName, err)
	}
	fmt.Printf("Update: %s installed sucessfully - pid:  %v.\n", newVersionName, p.Pid)
	// Create a context that will expire in 5 seconds and use this as a
	// timeout to Shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Return any errors during shutdown.
	err = httpServer.Shutdown(ctx)
	if err != nil {
		fmt.Printf("Error while shutting down server %v", err)
	}
	return httpServer
}

// Download file from local storage dir 'Dist'
// Should be passed an interface io.Reader for testing
// No error returned on Donwload
func DownloadAndVerifyFile(filePath string) error {

	// Open the file that should be copied
	// Read the contents
	// Create and open the file that the contents should be copied into
	// Write to the new file
	// Close both files
	fmt.Println(filePath)
	from, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer from.Close()

	execName, err := os.Executable()
	if err != nil {
		return err
	}

	err = VerifyRSASignature(filePath, filePath+SignatureExtension)
	if err != nil {
		return err
	}
	fmt.Println("Executable verified and Downloaded")

	os.Remove(execName)
	// A FileMode represents a file's mode and permission bits. 770 - Owner and Group have all, and Other can read and execute
	to, err := os.OpenFile(execName, os.O_RDWR|os.O_CREATE, 0770)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}

	return nil
}

// TODO: pass as an argument HARDCODED DIR '/DIST' PATH
// Check dir exists
// Assumption take first update sorted by date

// Make purposely no distinction between eventual
// local storage read errors and no file found
func checkNewVersion() string {
	for _, f := range listDir() {
		if fn := strings.Split(f, "."); len(fn) == 2 {
			newVersionName = f
			return fn[1]
		}
	}
	return ""
}

// Keep the eventual error inside the scope to not
// and unveil possible local storage error to http client
func listDir() (filesName []string) {
	files, err := ioutil.ReadDir(UpdateDir)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, f := range files {
		filesName = append(filesName, f.Name())
	}
	return
}

func main() {
	// Parse command line flags for the address to listen on.
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
