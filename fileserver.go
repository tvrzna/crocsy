package main

import (
	"fmt"
	"html"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type FileServer struct {
	r *Route
}

type MemoryUnit byte

func (b MemoryUnit) String() string {
	return []string{"B", "K", "M", "G", "T", "P"}[int(b)]
}

const (
	defaultIndex = "index.html"

	UnitB MemoryUnit = iota + 1
	UnitK
	UnitM
	UnitG
	UnitT
)

func newFileServer(r *Route) *FileServer {
	return &FileServer{r}
}

func (fs *FileServer) Serve(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path[len(fs.r.Path):], "/")
	if path == "" {
		path = "/"
	}

	filePath := filepath.Join(fs.r.Root, path)

	if fi, err := os.Stat(filePath); err == nil {
		if fi.IsDir() {
			if !strings.HasSuffix(path, "/") {
				http.Redirect(w, req, req.URL.Path+"/", http.StatusPermanentRedirect)
			}
			fs.serverDir(path, filePath, w)
		} else {
			fs.serverFile(filePath, fi.Size(), w)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (fs *FileServer) serverDir(path, filePath string, w http.ResponseWriter) {
	entries, err := os.ReadDir(filePath)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	indexName := defaultIndex
	if fs.r.Index != "" {
		indexName = fs.r.Index
	}
	indexName = strings.ToLower(indexName)
	indexNames := strings.Split(indexName, " ")

	for _, entry := range entries {
		if !entry.IsDir() {
			entryName := strings.ToLower(entry.Name())
			for _, in := range indexNames {
				if entryName == in {
					fi, err := entry.Info()
					if err != nil {
						log.Print(err)
						continue
					}
					fs.serverFile(filepath.Join(filePath, entry.Name()), fi.Size(), w)
					return
				}
			}
		}
	}

	if fs.r.Autoindex {
		fs.indexDir(entries, w, path)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (fs *FileServer) indexDir(entries []os.DirEntry, w http.ResponseWriter, path string) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}

		return entries[i].Name() < entries[j].Name()
	})

	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.Write([]byte("<!doctype html!><html><body><h1>" + html.EscapeString(path) + "</h1><table><tr><th>Name</th><th>Last Modified</th><th>Size</th></tr>"))
	if path != "/" {
		w.Write([]byte(fmt.Sprintf("<tr><td colspan=\"3\"><a href=\"%s\">../</a></td></tr>", "../")))
	}
	for _, entry := range entries {
		fi, err := entry.Info()
		if err != nil {
			log.Print(err)
			continue
		}
		lastModified := fi.ModTime().Format(("2006-01-02 15:04:05"))
		size := "<DIR>"
		link := filepath.Join(fs.r.Path, path, entry.Name())

		if entry.IsDir() {
			if !strings.HasSuffix(link, "/") {
				link += "/"
			}
		} else {
			size = tidySizePrefix(float64(fi.Size()), 0)
		}

		w.Write([]byte(fmt.Sprintf("<tr><td><a href=\"%s\">%s</a></td><td>%s</td><td>%s</td></tr>", html.EscapeString(link), html.EscapeString(entry.Name()), html.EscapeString(lastModified), html.EscapeString(size))))
	}
	w.Write([]byte("</table></body></html>"))
}

func (fs *FileServer) serverFile(filePath string, fileSize int64, w http.ResponseWriter) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fileName := filepath.Base(filePath)
	extension := ""
	if lastIndex := strings.LastIndex(fileName, "."); lastIndex > 0 {
		extension = fileName[lastIndex:]
	}

	mimeType := mime.TypeByExtension(extension)
	if mimeType == "" {
		mimeType = "text/plain"
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Length", strconv.FormatFloat(float64(fileSize), 'f', -1, 64))

	buffer := make([]byte, 1024)
	for {
		read, err := f.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Print(err)
			}
			break
		}
		w.Write(buffer[:read])
	}

}

func tidySizePrefix(value float64, start byte) string {
	size := value
	unit := MemoryUnit(start)
	for i := start; i < 5; i++ {
		if val := size / 1024; val < 1 {
			break
		} else {
			size = val
			unit = MemoryUnit(i + 1)
		}
	}

	return fmt.Sprintf("%.0f%s", size, unit.String())
}
