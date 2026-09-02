package main

import (
    "fmt"
    "io"
    "log"
    "mime/multipart"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

const uploadDir = "/data"

func main() {
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/upload", uploadHandler)
    http.HandleFunc("/files/", fileHandler)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintln(w, "ok")
    })

    log.Println("server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprint(w, `<!doctype html>
<html>
<head><title>Go Volume Demo</title></head>
<body>
<h2>Upload a file</h2>
<form action="/upload" method="post" enctype="multipart/form-data">
  <input type="file" name="file" required>
  <button type="submit">Upload</button>
</form>
<p>After upload, open the returned /files/&lt;name&gt; URL.</p>
</body>
</html>`)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "POST only", http.StatusMethodNotAllowed)
        return
    }

    if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, "invalid multipart form", http.StatusBadRequest)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "file field is required", http.StatusBadRequest)
        return
    }
    defer file.Close()

    name := safeFilename(header)
    dstPath := filepath.Join(uploadDir, name)

    dst, err := os.Create(dstPath)
    if err != nil {
        http.Error(w, "cannot create destination file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, "cannot save file", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    fmt.Fprintf(w, "uploaded: %s\nfetch: /files/%s\n", name, name)
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
    name := strings.TrimPrefix(r.URL.Path, "/files/")
    name = filepath.Base(name)
    if name == "." || name == "" || name == "/" {
        http.Error(w, "file name required", http.StatusBadRequest)
        return
    }

    path := filepath.Join(uploadDir, name)
    if _, err := os.Stat(path); err != nil {
        if os.IsNotExist(err) {
            http.NotFound(w, r)
            return
        }
        http.Error(w, "cannot read file", http.StatusInternalServerError)
        return
    }

    http.ServeFile(w, r, path)
}

func safeFilename(header *multipart.FileHeader) string {
    name := filepath.Base(header.Filename)
    name = strings.ReplaceAll(name, " ", "_")
    if name == "." || name == "" {
        return "upload.bin"
    }
    return name
}
