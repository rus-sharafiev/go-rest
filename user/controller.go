package user

import (
	"net/http"
	"strings"

	"github.com/rus-sharafiev/go-rest-common/db"
	"github.com/rus-sharafiev/go-rest-common/exception"
)

type controller struct {
	db *db.Postgres
}

func (c controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	switch parts := strings.Split(path, "/")[2:]; len(parts) {

	case 0:
		switch r.Method {
		case http.MethodPost:
			c.create(w, r)
		case http.MethodGet:
			c.findAll(w, r)
		default:
			exception.MethodNotAllowed(w)
		}

	case 1:
		id := parts[0]
		switch r.Method {
		case http.MethodGet:
			c.findOne(id, w, r)
		case http.MethodPatch:
			c.update(id, w, r)
		case http.MethodDelete:
			c.delete(id, w, r)
		default:
			exception.MethodNotAllowed(w)
		}

	default:
		http.NotFound(w, r)
	}
}

var Controller = &controller{db: &db.Instance}
