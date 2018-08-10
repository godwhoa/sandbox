package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"bitbucket.org/atlassian/limitbytes"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
)

const (
	// Kilo
	K = 1000
	// Mega
	M = 1000000
	// 5KB file upload limit
	MAX_FILE_SIZE = 5 * K
	// Execution timeout of 30 sec
	EXEC_TIMEOUT = 30
	// Memory limit of 100MB
	MAX_MEM = 100 * M
	// Pid limit
	MAX_PID = 100
	// CPU percentage per container
	CPU_PERCENT = 10
)

// Runner runs a given source file and returns stdout and stderr
type Runner interface {
	Run(ctx context.Context, srcfile string) (stdout, stderr []byte, err error)
}

type Response struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

// JSON response helper
func JSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if json.NewEncoder(w).Encode(v) != nil {
		http.Error(w, "Error occured while encoding JSON", http.StatusInternalServerError)
	}
}

func RunHandler(runner Runner) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		file, _, err := r.FormFile("src")
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// copy to tmp file
		tmpfile, err := ioutil.TempFile("", "sandbox")
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		defer tmpfile.Close()
		defer os.Remove(tmpfile.Name())
		io.Copy(tmpfile, file)

		ctx, cancel := context.WithTimeout(r.Context(), EXEC_TIMEOUT*time.Second)
		defer cancel()

		stdout, stderr, err := runner.Run(ctx, tmpfile.Name())
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, fmt.Sprintf("Execution time exceeded %dsec limit", EXEC_TIMEOUT), http.StatusRequestTimeout)
			return
		}

		JSON(w, Response{
			Stdout: string(stdout),
			Stderr: string(stderr),
		})

	}
}

func main() {
	r := chi.NewRouter()
	runner := NewDockerRunner()

	r.Use(middleware.Logger)
	r.Use(limitbytes.New(MAX_FILE_SIZE, func(w http.ResponseWriter, r *http.Request, e error) {
		switch e.(type) {
		case nil:
			w.WriteHeader(http.StatusOK)
		case limitbytes.ErrTooLarge:
			http.Error(w, "File size limit of 5KB exceeded", http.StatusRequestEntityTooLarge)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	r.Post("/run", RunHandler(runner))
	FileServer(r, "/", http.Dir("../frontend"))

	http.ListenAndServe(":8080", r)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
