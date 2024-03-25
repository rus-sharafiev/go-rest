package images

import (
	"net/http"

	"github.com/rus-sharafiev/go-rest-common/db"
	"github.com/rus-sharafiev/go-rest-common/exception"
)

type controller struct {
	db *db.Postgres
}

func (c controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		c.handle(w, r)

	case http.MethodGet:
		c.serve(w, r)

	default:
		exception.MethodNotAllowed(w)
	}
}

var Controller = &controller{db: &db.Instance}
