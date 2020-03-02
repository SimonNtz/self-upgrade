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
	Status         = struct{ Version, NewVersion string }{Version, ""}
	NewVersionName string
)

func (sh *specifiHandler) handler(w http.ResponseWriter, r *http.Request) {
	if err := sh.page.Execute(w, Status); err != nil {
		log.Fatal(err)
	}
}

// handlerCheck checks if a new version exists in local storage directory
func (sh *specifiHandler) handlerCheck(w http.ResponseWriter, r *http.Request) {

	if newVersion := checkNewVersion(); newVersion != "" {
		Status.NewVersion = newVersion
	}
	if err := sh.page.Execute(w, Status); err != nil {
		log.Fatal(err)
	}
}

// handlerInstall installs new version of current executable from local storage
func (sh *specifiHandler) handlerInstall(w http.ResponseWriter, r *http.Request) {

	// This condition protects our handler against attack by URL manipulation
	if checkNewVersion() == "" {
		http.Error(w, `Unauthorized access`, http.StatusUnauthorized)
		return
	}

	// Download and signature verificaiton of new executable starts here
	err := downloadAndVerifyFile(filepath.Join(UpdateDir, NewVersionName))
	if err != nil {
		errString := fmt.Sprintf("Error while donwloading %s: %v", NewVersionName, err)
		log.Fatalf(fmt.Sprintf("Error while donwloading %s: %v", NewVersionName, err))
		http.Error(w, errString, http.StatusInternalServerError)
		return
	}

	// Redirects client's browser to home page before server shutdown
	defer http.Redirect(w, r, "/", 302)
	sh.signalCh <- syscall.SIGINT

}

// Start server and wait for the SIGINT signal to restart
//  with the updated version while preserging the TCP socket
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

	// Restart the with updated executable while preserving the socket info
	p, err := RestartExec(addr, ln)
	if err != nil {
		fmt.Printf("Error while installing update: %s: %v.\n", NewVersionName, err)
	}
	fmt.Printf("Update: %s installed sucessfully - pid:  %v.\n", NewVersionName, p.Pid)
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

// Download and verifiy the new executable file
// from local storage dir "Dist"
func downloadAndVerifyFile(filePath string) error {

	from, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer from.Close()

	execName, err := os.Executable()
	if err != nil {
		return err
	}

	// Verifiy file's RSA signature before removing current executable
	err = VerifyRSASignature(filePath, filePath+SignatureExtension)
	if err != nil {
		return err
	}
	fmt.Println("Executable verified and Downloaded")

	// Remove current executable file
	os.Remove(execName)

	to, err := os.OpenFile(execName, os.O_RDWR|os.O_CREATE, 0770)
	if err != nil {
		return err
	}
	defer to.Close()

	// Replace current executable
	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}

	return nil
}

// Make purposely no distinction between eventual
// local storage read error and no file found
func checkNewVersion() string {
	for _, f := range listDir() {
		if fn := strings.Split(f, "."); len(fn) == 2 {
			NewVersionName = f
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

	fmt.Printf("Exiting.\n")
}
