package uploads

import (
	"net/http"
	"path/filepath"

	"github.com/rus-sharafiev/go-rest/common"
	"github.com/rus-sharafiev/go-rest/common/auth"
	"github.com/rus-sharafiev/go-rest/common/exception"
)

type handler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		exception.MethodNotAllowed(w)
		return
	}
	userId, role := auth.Headers(r)
	if len(userId) == 0 {
		exception.Unauthorized(w)
		return
	}

	uploadDir := filepath.Join(*common.Config.StaticDir, *common.Config.UploadDir, userId)
	if role == "ADMIN" {
		uploadDir = filepath.Join(*common.Config.StaticDir, *common.Config.UploadDir)
	}

	w.Header().Add("Cache-Control", "private, max-age=31536000, immutable")
	http.StripPrefix(*common.Config.UploadPath, http.FileServer(http.Dir(uploadDir))).ServeHTTP(w, r)
}

var Handler = &handler{}
