package images

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rus-sharafiev/go-rest-common/exception"
)

// HANDLE UPLOAD ------------------------------------------------------------------
func (c controller) upload(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		exception.InternalServerError(w, err)
	}

	w.Write(body)
}

// HANDLE REMOVE ------------------------------------------------------------------
func (c controller) remove(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	name := r.PathValue("name")

	go func() {
		if len(name) > 0 {
			basename := filepath.Base(name)
			name := strings.TrimSuffix(basename, filepath.Ext(basename))

			files, err := filepath.Glob(".static/images/" + name + "*")
			if err != nil {
				exception.InternalServerError(w, err)
				return
			}

			for _, file := range files {
				if err := os.Remove(file); err != nil {
					fmt.Println(err)
				}
			}
		}
	}()

	result := Message{
		StatusCode: http.StatusOK,
		Message:    "",
	}

	// OK response
	json.NewEncoder(w).Encode(&result)
}
