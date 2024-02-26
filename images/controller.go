package images

import (
	"net/http"

	"github.com/rus-sharafiev/go-rest/common/exception"
)

type Controller struct {
	UploadDir string
	SubmitUrl string
}

func (c Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		c.handle(w, r)

	case http.MethodGet:
		c.serve(w, r)

	default:
		exception.MethodNotAllowed(w)
	}
}
