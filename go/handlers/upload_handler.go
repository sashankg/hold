package handlers

import (
	"encoding/base64"
	"hash"
	"hash/fnv"
	"io"
	"net/http"
	"os"

	"github.com/sashankg/hold/server"
)

type UploadHandler struct {
	hasher Hasher
}

func NewUploadHandler(
	hasher Hasher,
) *UploadHandler {
	return &UploadHandler{
		hasher,
	}
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "../server/static/upload.html")
	case http.MethodPost:
		reader, err := r.MultipartReader()
		if err != nil {
			server.InternalServerError(w, err)
			return
		}
		for {
			part, err := reader.NextPart()
			if err != nil {
				if err == io.EOF {
					break
				}
				server.InternalServerError(w, err)
				return
			}
			println(part.FileName(), part.FormName(), part.Header)
			file, err := os.Create("uploads/" + part.FileName())
			defer file.Close()

			hash := h.hasher.New()

			mw := io.MultiWriter(hash, file)

			io.Copy(mw, part)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(base64.StdEncoding.EncodeToString(hash.Sum(nil))))
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *UploadHandler) Route() string {
	return "/upload"
}

type Hasher interface {
	New() hash.Hash
}

type FnvHasher struct{}

func (f *FnvHasher) New() hash.Hash {
	return fnv.New128()
}
