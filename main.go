package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/unixpickle/essentials"
)

func main() {
	var addr string
	var filename string
	flag.StringVar(&addr, "addr", ":8080", "the address to listen on")
	flag.StringVar(&filename, "filename", "", "the name of the file to serve up")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: dreamcatcher [flags] <endpoint URL>")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		essentials.Die()
	}

	endpointStr := flag.Args()[0]
	endpoint, err := url.Parse(endpointStr)
	essentials.Must(err)

	reader, err := NewHTTPReader(endpoint)
	essentials.Must(err)
	cache := NewDataCache(int(reader.Size))

	if filename == "" {
		filename = reader.Name
	}
	log.Println("Using filename:", filename)

	time := time.Now()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("range") != "" {
			log.Println("request at range", r.Header.Get("range"))
		} else {
			log.Println("request with no range")
		}
		options := map[string]string{"filename": filename}
		w.Header().Set("content-disposition", mime.FormatMediaType("attachment", options))
		f := &ErrorLogger{NewCachedReader(reader.Dup(), cache)}
		http.ServeContent(w, r, filename, time, f)
	})

	log.Println("Listening on address", addr, "...")
	http.ListenAndServe(addr, nil)
}

type ErrorLogger struct {
	io.ReadSeeker
}

func (e *ErrorLogger) Seek(offset int64, whence int) (int64, error) {
	off, err := e.ReadSeeker.Seek(offset, whence)
	if err != nil {
		log.Println("seek error:", err)
	}
	return off, err
}

func (e *ErrorLogger) Read(p []byte) (int, error) {
	n, err := e.ReadSeeker.Read(p)
	if err != nil {
		log.Println("read error:", err)
	}
	return n, err
}
