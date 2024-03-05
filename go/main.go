package main

import (
	"io"
	"net"
	"net/http"
	"os"
)

func main() {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: 3000,
	})
	if err != nil {
		panic(err)
	}
	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.ServeFile(w, r, "../server/static/upload.html")
		case http.MethodPost:
			reader, err := r.MultipartReader()
			if err != nil {
				InternalServerError(w, err)
				return
			}
			for {
				part, err := reader.NextPart()
				if err != nil {
					if err == io.EOF {
						break
					}
					InternalServerError(w, err)
					return
				}
				println(part.FileName(), part.FormName(), part.Header)
				file, err := os.Create("uploads/" + part.FileName())
				defer file.Close()

				file.ReadFrom(part)
				w.WriteHeader(http.StatusOK)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.Serve(ln, serveMux)
}

func InternalServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}
