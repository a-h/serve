package handlers

import (
	"bytes"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileHandler(t *testing.T) {
	// Create a temporary directory for testing
	dir, err := os.MkdirTemp("", "filehandler_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create a test file in the temporary directory.
	testFilePath := filepath.Join(dir, "testfile.txt")
	err = os.WriteFile(testFilePath, []byte("Hello, World!"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a new FileHandler, in read only mode.
	log := slog.New(slog.DiscardHandler)
	fh, closer, err := NewFileHandler(log, dir, true)
	if err != nil {
		t.Fatalf("Failed to create FileHandler: %v", err)
	}
	defer closer()

	t.Run("Existing files are returned", func(t *testing.T) {
		testGet(t, fh, "/testfile.txt", http.StatusOK, "Hello, World!")
	})
	t.Run("Non-existing files return 404", func(t *testing.T) {
		testGet(t, fh, "/nonexistent.txt", http.StatusNotFound, "404 page not found\n")
	})
	t.Run("Cannot PUT when not writable", func(t *testing.T) {
		testWrite(t, fh, http.MethodPut, "/newfile.txt", "New content", http.StatusMethodNotAllowed)
	})
	t.Run("Cannot POST when not writable", func(t *testing.T) {
		testWrite(t, fh, http.MethodPost, "/newfile.txt", "New content", http.StatusMethodNotAllowed)
	})
	t.Run("Cannot DELETE when not writable", func(t *testing.T) {
		testWrite(t, fh, http.MethodDelete, "/testfile.txt", "", http.StatusMethodNotAllowed)
	})

	// Update the FileHandler to be writable and test writing a new file.
	fh.IsReadOnly = false
	t.Run("Can PUT when writable", func(t *testing.T) {
		testWrite(t, fh, http.MethodPut, "/newfile.txt", "New content", http.StatusCreated)
		testGet(t, fh, "/newfile.txt", http.StatusOK, "New content")
	})
	t.Run("Can POST when writable", func(t *testing.T) {
		testWrite(t, fh, http.MethodPost, "/anotherfile.txt", "Another content", http.StatusCreated)
		testGet(t, fh, "/anotherfile.txt", http.StatusOK, "Another content")
	})
	t.Run("Can DELETE when writable", func(t *testing.T) {
		testWrite(t, fh, http.MethodDelete, "/testfile.txt", "", http.StatusNoContent)
		testGet(t, fh, "/testfile.txt", http.StatusNotFound, "404 page not found\n")
	})
	t.Run("Cannot write outside root directory", func(t *testing.T) {
		testWrite(t, fh, http.MethodPut, "/..//outside.txt", "Should not be created outside", http.StatusCreated)
		if _, err := os.Stat(filepath.Join(dir, "../outside.txt")); !os.IsNotExist(err) {
			t.Errorf("File was created outside root directory")
		}
		testGet(t, fh, "/outside.txt", http.StatusOK, "Should not be created outside")
	})
	t.Run("Can create files in subdirectories", func(t *testing.T) {
		testWrite(t, fh, http.MethodPut, "/subdir/newfile.txt", "Subdirectory content", http.StatusCreated)
		testGet(t, fh, "/subdir/newfile.txt", http.StatusOK, "Subdirectory content")
	})
	t.Run("Can upload multipart form file", func(t *testing.T) {
		testMultipartUpload(t, fh, "/multipart.txt", "Multipart content", http.StatusCreated)
		testGet(t, fh, "/multipart.txt", http.StatusOK, "Multipart content")
	})
}

func testGet(t *testing.T, fh *FileHandler, urlPath string, expectedStatus int, expectedBody string) {
	req := httptest.NewRequest(http.MethodGet, urlPath, nil)
	w := httptest.NewRecorder()
	fh.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
	defer resp.Body.Close()
	actualBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(actualBytes) != expectedBody {
		t.Errorf("Expected response body %q, got %q", expectedBody, string(actualBytes))
	}
}

func testWrite(t *testing.T, fh *FileHandler, method string, urlPath string, body string, expectedStatus int) {
	req := httptest.NewRequest(method, urlPath, io.NopCloser(strings.NewReader(body)))
	w := httptest.NewRecorder()
	fh.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}

func testMultipartUpload(t *testing.T, fh *FileHandler, urlPath string, fileContent string, expectedStatus int) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "upload.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, err = part.Write([]byte(fileContent))
	if err != nil {
		t.Fatalf("Failed to write to form file: %v", err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, urlPath, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	fh.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}
