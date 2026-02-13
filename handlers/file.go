package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

func NewFileHandler(dir string, readOnly bool) (fh *FileHandler, closer func() error, err error) {
	fh = &FileHandler{
		IsReadOnly: readOnly,
		fileServer: http.FileServer(http.Dir(dir)),
	}
	fh.rootedFileSystem, err = os.OpenRoot(dir)
	if err != nil {
		return fh, nil, fmt.Errorf("failed to open root directory: %w", err)
	}
	closer = func() error {
		return fh.rootedFileSystem.Close()
	}
	return fh, closer, nil
}

type FileHandler struct {
	IsReadOnly       bool
	fileServer       http.Handler
	rootedFileSystem *os.Root
}

func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		h.Get(w, r)
	case http.MethodPost, http.MethodPut:
		h.Put(w, r)
	case http.MethodDelete:
		h.Delete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *FileHandler) Get(w http.ResponseWriter, r *http.Request) {
	h.fileServer.ServeHTTP(w, r)
}

func (h *FileHandler) cleanPath(p string) string {
	cleaned := path.Clean(p)
	if cleaned == "." || strings.Contains(cleaned, "..") {
		cleaned = ""
	}
	return strings.TrimPrefix(cleaned, "/")
}

func (h *FileHandler) Put(w http.ResponseWriter, r *http.Request) {
	if h.IsReadOnly {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cleaned := h.cleanPath(r.URL.Path)
	if cleaned == "" {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}
	if r.Body == nil {
		http.Error(w, "No file content provided", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Create the file.
	err := h.rootedFileSystem.MkdirAll(path.Dir(cleaned), 0755)
	if err != nil {
		fmt.Printf(" - failed to create directories: %v\n", err)
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	f, err := h.rootedFileSystem.Create(cleaned)
	if err != nil {
		fmt.Printf(" - failed to create file: %v\n", err)
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// Read the file content from the request body.
	reader, err := h.getReader(r)
	if err != nil {
		fmt.Printf(" - failed to get file reader: %v\n", err)
		http.Error(w, "failed to read file content", http.StatusBadRequest)
		return
	}

	_, err = f.ReadFrom(reader)
	if err != nil {
		fmt.Printf(" - failed to write file: %v\n", err)
		http.Error(w, "failed to write file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *FileHandler) getReader(r *http.Request) (io.Reader, error) {
	isMultipart := strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data")
	if !isMultipart {
		return r.Body, nil
	}
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}
loop:
	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break loop
			}
			return nil, err
		}
		if part.FormName() == "file" {
			return part, nil
		}
	}
	return nil, fmt.Errorf("file part not found in multipart form data")
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if h.IsReadOnly {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := h.rootedFileSystem.Remove(h.cleanPath(r.URL.Path))
	if err != nil {
		fmt.Printf(" - failed to delete file: %v\n", err)
		http.Error(w, "failed to delete file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
