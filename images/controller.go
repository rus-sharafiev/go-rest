package images

import (
	"net/http"

	"github.com/rus-sharafiev/go-rest-common/db"
)

type controller struct {
	db *db.Postgres
}

func (c controller) Handler(mux *http.ServeMux) {
	mux.HandleFunc("POST /images", c.upload)
	mux.HandleFunc("POST /images/{$}", c.upload)

	mux.HandleFunc("DELETE /images/{name}", c.remove)
}

var Controller = &controller{db: &db.Instance}
