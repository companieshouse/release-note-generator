package main

import (
	"fmt"
	"github.com/release-note-generator/pre"
	"log"
	"net/http"
	"os/exec"
	"runtime"
)

var w http.ResponseWriter
var r *http.Request

func main() {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	http.Handle("/ch_logo.jpg", http.FileServer(http.Dir("./ch_logo.jpg")))
	http.HandleFunc("/", ServeTemplate)

	open("http://localhost:8080/template.htm")
	fmt.Println("Listening...")
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		return
	}
}

func ServeTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	pre.ServeTemplatePre(w, r)
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
