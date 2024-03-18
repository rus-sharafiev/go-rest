package spa

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rus-sharafiev/go-rest/common"
	"github.com/rus-sharafiev/go-rest/common/exception"
)

type handler struct{}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		exception.BadRequestError(w, err)
		return
	}

	path = filepath.Join(*common.Config.StaticDir, path)
	index := "index.html"

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(*common.Config.StaticDir, index))
		return
	} else if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	// Serve gziped
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {

		if fileType := filepath.Ext(path); fileType == ".js" || fileType == ".css" {
			if _, err := os.Stat(path + ".gz"); err == nil {

				w.Header().Add("Content-Encoding", "gzip")
				if fileType == ".js" {
					w.Header().Add("Content-Type", "application/javascript")
				}
				if fileType == ".css" {
					w.Header().Add("Content-Type", "text/css")
				}
				http.ServeFile(w, r, filepath.Join(path+".gz"))

				return
			}
		}
	}

	w.Header().Add("Cache-Control", "no-cache")
	http.FileServer(http.Dir(*common.Config.StaticDir)).ServeHTTP(w, r)
}

var Handler = &handler{}
