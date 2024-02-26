package uploads

import (
	"fmt"
	"net/http"

	"github.com/rus-sharafiev/go-rest/common/exception"
)

type Server struct {
	Dir *string
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		exception.MethodNotAllowed(w)
		return
	}

	fmt.Printf("requested file is: %s", r.URL.Path)

	// TODO check user id, to serve private folder

	w.Header().Add("Cache-Control", "private, max-age=31536000, immutable")
	http.FileServer(http.Dir(*s.Dir)).ServeHTTP(w, r)
}
