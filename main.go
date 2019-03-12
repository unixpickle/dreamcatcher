package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/unixpickle/essentials"
)

func main() {
	var filename string

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

	if filename == "" {
		filename = path.Base(endpoint.Path)
		// TODO: find basename from content disposition.
	}
}
