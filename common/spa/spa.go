package spa

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/rus-sharafiev/go-rest/common/exception"
)

type Handler struct {
	Static *string
	Index  *string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		exception.BadRequestError(w, err)
		return
	}

	path = filepath.Join(*h.Static, path)
	index := "index.html"
	if h.Index != nil {
		index = *h.Index
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(*h.Static, index))
		return
	} else if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	w.Header().Add("Cache-Control", "no-cache")
	http.FileServer(http.Dir(*h.Static)).ServeHTTP(w, r)
}
