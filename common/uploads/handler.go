package uploads

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/rus-sharafiev/go-rest/common"
	"github.com/rus-sharafiev/go-rest/common/exception"
	"github.com/rus-sharafiev/go-rest/common/jwt"
)

type handler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		exception.MethodNotAllowed(w)
		return
	}

	token := r.URL.Query().Get("token")
	if len(token) == 0 {
		exception.Unauthorized(w)
		return
	}

	claims, err := jwt.Validate(token)
	if err != nil {
		exception.BadRequestMessage(w, err.Error())
		return
	}

	uploadDir := filepath.Join(*common.Config.UploadDir, strconv.Itoa(claims.UserId))
	// if claims.UserAccess == "ADMIN" {
	// 	uploadDir = filepath.Join(*common.Config.UploadDir)
	// }

	r.URL.RawQuery = ""
	w.Header().Add("Cache-Control", "private, max-age=31536000, immutable")
	http.StripPrefix(*common.Config.UploadPath, http.FileServer(http.Dir(uploadDir))).ServeHTTP(w, r)
}

var Handler = &handler{}
