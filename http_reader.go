package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type HTTPReader struct {
	Size     int64
	Endpoint *url.URL
	Name     string

	offset int64
}

func NewHTTPReader(u *url.URL) (*HTTPReader, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	length := resp.Header.Get("content-length")
	size, err := strconv.ParseInt(length, 10, 64)
	if err != nil {
		return nil, err
	}
	name := path.Base(u.Path)
	if resp.Header.Get("content-disposition") != "" {
		_, params, err := mime.ParseMediaType(resp.Header.Get("content-disposition"))
		if err != nil {
			return nil, err
		}
		if params["filename"] != "" {
			name = params["filename"]
		}
	}
	return &HTTPReader{
		Size:     size,
		Endpoint: u,
		Name:     name,
	}, nil
}

func (r *HTTPReader) Dup() *HTTPReader {
	return &HTTPReader{
		Size:     r.Size,
		Endpoint: r.Endpoint,
		Name:     r.Name,
	}
}

func (r *HTTPReader) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekCurrent {
		offset += r.offset
	} else if whence == io.SeekEnd {
		offset += r.Size
	}
	if offset < 0 {
		return 0, errors.New("negative seek offset")
	} else if offset > r.Size {
		offset = r.Size
	}
	r.offset = offset
	return offset, nil
}

func (r *HTTPReader) Read(p []byte) (int, error) {
	if r.offset == r.Size {
		return 0, io.EOF
	}
	size := int64(len(p))
	if r.Size-r.offset < size {
		size = r.Size - r.offset
	}
	req, err := http.NewRequest("GET", r.Endpoint.String(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("range", fmt.Sprintf("bytes=%d-%d", r.offset, r.offset+size))
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	n, err := io.ReadFull(response.Body, p)
	r.offset += int64(n)
	return n, err
}
